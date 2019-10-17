package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/kyokomi/emoji"
	"github.com/rcompos/go-bitbucket"
)

//const sleepytime time.Duration = 2

func main() {

	rand.Seed(time.Now().UnixNano())

	var bbUser, bbPassword, bbOwner, bbRole, searchStr, replaceStr, inFile, outFile, branch, prTitle string
	var execute, createPR, listOnly bool
	//var numWorkers int

	flag.StringVar(&bbUser, "u", os.Getenv("BITBUCKET_USERNAME"), "Bitbucket user (required) (envvar BITBUCKET_USERNAME)")
	flag.StringVar(&bbPassword, "p", os.Getenv("BITBUCKET_PASSWORD"), "Bitbucket password (required) (envvar BITBUCKET_PASSWORD)")
	flag.StringVar(&bbOwner, "o", os.Getenv("BITBUCKET_OWNER"), "Bitbucket owner (required) (envvar BITBUCKET_OWNER)")
	flag.StringVar(&bbRole, "e", os.Getenv("BITBUCKET_ROLE"), "Bitbucket role (envvar BITBUCKET_ROLE)")
	flag.StringVar(&searchStr, "s", os.Getenv("BITBUCKET_SEARCH"), "Text to search for (envvar BITBUCKET_SEARCH)")
	flag.StringVar(&replaceStr, "r", os.Getenv("BITBUCKET_REPLACE"), "Text to replace with (envvar BITBUCKET_REPLACE)")
	flag.StringVar(&branch, "b", os.Getenv("BITBUCKET_BRANCH"), "Feature branch where changes are made (envvar BITBUCKET_BRANCH)")
	flag.StringVar(&prTitle, "t", os.Getenv("BITBUCKET_TITLE"), "Title for pull request (envvar BITBUCKET_PRTITLE)")
	flag.StringVar(&inFile, "i", "", "Input file")
	flag.StringVar(&outFile, "f", "./logs/out.txt", "Output file")
	flag.BoolVar(&execute, "x", false, "Execute text replace")
	flag.BoolVar(&createPR, "c", false, "Create pull request")
	flag.BoolVar(&listOnly, "l", false, "Return repo list only")
	//flag.IntVar(&numWorkers, "w", 100, "Number of worker threads")
	flag.Parse()

	bbURL := "https://" + bbUser + ":" + bbPassword + "@bitbucket.org"
	var repoCache []string

	if bbUser == "" || bbPassword == "" || bbOwner == "" {
		fmt.Println("Must supply user (-u), password (-p) and owner (-o)!")
		fmt.Println("Alternately, environmental variables can be set.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if execute == true && (searchStr == "" || replaceStr == "") {
		fmt.Println("Must supply search string (-s) and replace string (-r)!")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if execute && !listOnly { // skip if listOnly
		promptRead(bbOwner, searchStr, replaceStr)
	}

	logsDir := "logs"
	createDir(logsDir)
	logFileDateFormat := "2006-01-02-150405"
	logStamp := time.Now().Format(logFileDateFormat)
	logfile := logsDir + "/bb-sar-" + string(logStamp) + ".log"

	logf, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		log.Fatal(err)
	}
	defer logf.Close()
	log.SetOutput(logf) //log.Println("Test log message")

	log.Printf("Current Unix Time: %v\n", time.Now().Unix())

	//fmt.Println(path.Base(os.Args[0]))
	//time.Sleep(2)

	opt := &bitbucket.RepositoriesOptions{}
	opt.Owner = bbOwner
	opt.Role = bbRole // [owner|admin|contributor|member]

	c := bitbucket.NewBasicAuth(bbUser, bbPassword)

	var repos *bitbucket.RepositoriesRes
	var repoList []string

	if _, err := os.Stat(inFile); err == nil && inFile != "" {
		// inFile exists, use it
		//fmt.Printf("inFile found! %v\n", inFile)
		readDiskCache(&repoCache, inFile)
		//reps := []string{}
		//reps = repoCache
		//fmt.Printf("Repos:\n")
		for _, j := range repoCache {
			//fmt.Printf(">>  i: %v   j: %v\n", i, j)
			repoList = append(repoList, j)
		}
	} else {
		//  If inFile not exist, then request from BB API
		repos, err = c.Repositories.ListForAccount(opt)
		if err != nil {
			panic(err)
		}
		repositories := repos.Items
		for _, v := range repositories {
			repoList = append(repoList, v.Full_name)
		}
	}

	writeDiskCache(&repoList, outFile)

	if listOnly {
		for _, j := range repoList {
			fmt.Println(j)
		}
		os.Exit(0)
	}

	color.Set(color.FgMagenta)
	emoji.Printf("Acquiring repos for %s [ clone :hamburger: | pull :fries: | untracked :gem: | pull request :thumbsup: ]\n\n", bbOwner)
	color.Unset() // Don't forget to unset

	reposBaseDir := "repos"
	createDir(fmt.Sprintf("%s/%s", reposBaseDir, bbOwner))

	var wg sync.WaitGroup

	// TODO: Change from waitgroup to buffered channels

	//fmt.Printf("Repos:\n")
	for i, j := range repoList {
		wg.Add(1)
		//fmt.Printf("%v> %v\n", i, j)
		go func(i int, createPR bool, j, dir, owner, search, replace, user, pw, fBranch, pr, url string) {
			//go func(i int, j string, dir string, owner string, search string, replace string, createPR bool, user string, pw string) {
			defer wg.Done()
			dirOwner := dir + "/" + owner
			dirRepo := dir + "/" + j
			fmt.Printf("%s/%s\n", url, j)

			// add check for lines starting with #
			if strings.HasPrefix(j, "#") {
				fmt.Printf("Skipping: %v\n", j)
			}

			repoSCM := "curl --user " + user + ":" + pw + " https://api.bitbucket.org/2.0/repositories/" + j + `| jq | grep '\"scm\":' | perl -pe's/^\s*\"scm\": "(\S+)"\,\s*$/$1/'`
			//fmt.Printf("repoSCM: %v\n", repoSCM)
			repoSCMExec := exec.Command("bash", "-c", repoSCM)
			repoSCMExec.Dir = dirOwner
			repoSCMExecOut, _ := repoSCMExec.Output()
			scm := string(repoSCMExecOut)
			//fmt.Printf("SCM: '%v'\n", scm)

			if scm == "git" {

				gitClone := "git clone " + url + "/" + j
				errCloneNum := repoAction(j, gitClone, dirOwner, ":hamburger:", "", "", "")
				if errCloneNum == 128 {

					gitPullOrigin := "git pull origin `git symbolic-ref refs/remotes/origin/HEAD | sed 's@^refs/remotes/origin/@@' master`"
					errPullOriginNum := repoAction(j, gitPullOrigin, dirRepo, ":fire:", "", "", "")
					if errPullOriginNum != 0 {
						emoji.Printf(":fire:")
					}

					//gitPull := "git pull"
					gitPull := "git branch --set-upstream-to=origin/" + fBranch + " " + fBranch
					errPullNum := repoAction(j, gitPull, dirRepo, ":fries:", "", "", "")
					if errPullNum == 1 {
						upstreamCmd := "git push --set-upstream origin " + fBranch
						upstreamExec := exec.Command("bash", "-c", upstreamCmd)
						upstreamExec.Dir = dirRepo
						upstreamExec.Output()
						upstreamExecOut, _ := upstreamExec.Output()
						upstreamResult := string(upstreamExecOut)
						fmt.Printf(">> upstreamResult: %v\n", upstreamResult)
					}
					if errPullNum != 0 {
						emoji.Printf(":fork_and_knife:")
					}
				}

				branchCmd := "git checkout -b " + fBranch
				fmt.Printf("> %v\n", branchCmd)
				branchExec := exec.Command("bash", "-c", branchCmd)
				branchExec.Dir = dirRepo
				branchExecOut, _ := branchExec.Output()
				bResult := string(branchExecOut)
				fmt.Printf(bResult)

				upstreamCmd := "git push --set-upstream origin " + fBranch
				fmt.Printf("> %v\n", upstreamCmd)
				upstreamExec := exec.Command("bash", "-c", upstreamCmd)
				upstreamExec.Dir = dirRepo
				upstreamExec.Output()
				upstreamExecOut, _ := upstreamExec.Output()
				upstreamResult := string(upstreamExecOut)
				fmt.Printf(upstreamResult)

				var sar string
				//# One-liner shell command will search and replace all files recursively.
				if execute == true {
					sar = `find . -path ./.git -prune -o -type f -print  -exec grep -Iq . {} \; -exec perl -i -pe"s/` +
						search + `/` + replace + `/g" {} \;`
				} else if search != "" && replace != "" {
					sar = `find . -path ./.git -prune -o -type f -print -exec grep -Iq . {} \; -exec perl -ne" print if s/` +
						search + `/` + replace + `/g" {} \;`
				} else if search != "" {
					sar = `find . -path ./.git -prune -o -type f -print -exec grep -Iq . {} \; -exec perl -ne" print if /` +
						search + `/g" {} \;`
				}

				if sar != "" {
					fmt.Printf("> %v\n", sar)
					sarExec := exec.Command("bash", "-c", sar)
					sarExec.Dir = dirRepo
					sarExecOut, err := sarExec.Output()
					if err != nil {
						panic(err)
						fmt.Printf("ERROR: %v\n", err)
					}
					searchResult := string(sarExecOut)
					fmt.Printf(searchResult)
				}

				//fmt.Println("dirRepo: ", dirRepo)
				// Check for untracked changes
				gitDiffIndex := "git diff-index --quiet HEAD --"
				errPullNum := repoAction(j, gitDiffIndex, dirRepo, "", "", "", "")
				if errPullNum != 0 {
					// Git untracked changes exist
					emoji.Printf(":gem:")
					// create Pull Request

					if createPR == true {
						commitCmd := "git commit -am'Replace " + search + " with " + replace + "'"
						fmt.Printf("> %v\n", commitCmd)
						commitExec := exec.Command("bash", "-c", commitCmd)
						commitExec.Dir = dirRepo
						commitExec.Output()
						commitExecOut, _ := commitExec.Output()
						//if err != nil {
						//	panic(err)
						//	fmt.Printf("ERROR: %v\n", err)
						//}
						commitResult := string(commitExecOut)
						fmt.Printf(commitResult)

						pushCmd := "git push"
						fmt.Printf("> %v\n", pushCmd)
						pushExec := exec.Command("bash", "-c", pushCmd)
						pushExec.Dir = dirRepo
						pushExec.Output()
						pushExecOut, _ := pushExec.Output()
						//if err != nil {
						//	panic(err)
						//	fmt.Printf("ERROR: %v\n", err)
						//}
						pushResult := string(pushExecOut)
						fmt.Printf(pushResult)

						titlePR := pr
						curlPR := fmt.Sprintf("curl -v https://api.bitbucket.org/2.0/repositories/%s/pullrequests "+
							"-u %s:%s --request POST --header 'Content-Type: application/json' "+
							"--data '{\"title\": \"%s\", \"source\": { \"branch\": { \"name\": \"%s\" } } }'", j, user, pw, titlePR, fBranch)

						//fmt.Printf("curlPR:\n%s\n", curlPR)
						prExec := exec.Command("bash", "-c", curlPR)
						prExec.Dir = dirRepo
						_, err := prExec.Output()
						prExecOut, err := prExec.Output()
						if err == nil {
							emoji.Printf(":thumbsup:")
						}
						//if err != nil {
						//	panic(err)
						//	fmt.Printf("ERROR: %v\n", err)
						//}
						prResult := string(prExecOut)
						fmt.Printf("PR> %v\n", prResult)
					}

				}
			} else {
				fmt.Printf("ERROR: Unsupported SCM: %s %s\n", j, scm)
				log.Printf("ERROR: Unsupported SCM: %s %s\n", j, scm)
			}

			// end git
			//fmt.Printf("%vth goroutine done.\n", i)
		}(i, createPR, j, reposBaseDir, bbOwner, searchStr, replaceStr, bbUser, bbPassword, branch, prTitle, bbURL)
	}

	wg.Wait()
	//fmt.Printf("\n\nAll goroutines complete.")
	fmt.Println()

} //

func repoAction(r, cmdstr, rdir, win, any, fail, fcess string) int {

	cmd := exec.Command("bash", "-c", cmdstr)
	fmt.Printf("cmd> %s\n", cmd)
	cmd.Dir = rdir

	var errorNumber int = 0
	var waitStatus syscall.WaitStatus
	if err := cmd.Run(); err != nil {
		if err != nil {
			//os.Stderr.WriteString(fmt.Sprintf("Error: %s\n", err.Error()))
			log.Printf("Error: %s\n", err.Error())
			if fail != "" {
				emoji.Printf(fail)
			}
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			errorNumber = waitStatus.ExitStatus()
			log.Printf("Output: %s\n", []byte(fmt.Sprintf("%d", waitStatus.ExitStatus())))
			if fcess != "" {
				emoji.Printf(fcess)
			}
		}
	} else {
		// Success
		waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
		errorNumber = waitStatus.ExitStatus()
		log.Printf("Output: %s\n", []byte(fmt.Sprintf("%d", waitStatus.ExitStatus())))
		if win != "" {
			emoji.Printf(win)
		}
	}
	if any != "" {
		emoji.Printf(any)
	}
	return errorNumber
}

func createDir(dir string) {
	//	if [ -d "repos" ]; then echo true; else echo false; fi
	mkdirCmd := fmt.Sprintf("if [ ! -d %s ]; then mkdir -p -m775 %s; fi", dir, dir)
	mkdirExec := exec.Command("bash", "-c", mkdirCmd)
	mkdirExecOut, err := mkdirExec.Output()
	if err != nil {
		panic(err)
	}
	//fmt.Println(mkdirCmd)
	fmt.Printf(string(mkdirExecOut))

}

func checkFileInfo(f string) {
	//fi, err := os.Lstat("some-filename")
	fi, err := os.Lstat(f)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("permissions: %#o\n", fi.Mode().Perm()) // 0400, 0777, etc.
	switch mode := fi.Mode(); {
	case mode.IsRegular():
		fmt.Println("regular file")
	case mode.IsDir():
		fmt.Println("directory")
	case mode&os.ModeSymlink != 0:
		fmt.Println("symbolic link")
	case mode&os.ModeNamedPipe != 0:
		fmt.Println("named pipe")
	}
}

func promptRead(owner string, s string, r string) {
	reader := bufio.NewReader(os.Stdin)
	if s == "" {
		fmt.Printf("Git clone all %s repos.\n", owner)
	} else {
		fmt.Printf("Perform text replacement in all %s repos.\n", owner)
		fmt.Printf("%s -> %s\n", s, r)
	}
	fmt.Printf("To continue type '%s': \n", owner)
	text, _ := reader.ReadString('\n')
	answer := strings.TrimRight(text, "\n")
	//fmt.Printf("answer: %s \n", answer)
	//if answer == "y" || answer == "Y" {
	if answer == owner {
		return
	} else {
		//prompt2() //For recursive prompting
		log.Fatal("Exiting without action.")
	}
}

func readDiskCache(c *[]string, cf string) {
	//fmt.Println("Reading cache from disk: ", cf)
	var lines []string
	lines = readInFile(cf)
	for _, v := range lines {
		(*c) = append(*c, v)
	}
}

func writeDiskCache(c *[]string, cf string) {
	// If the file doesn't exist, create it, or append to the file
	//f, err := os.OpenFile(cf, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	f, err := os.OpenFile(cf, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	for _, w := range *c {
		outString := fmt.Sprintf("%v\n", w)
		if _, err := f.WriteString(outString); err != nil {
			log.Println(err)
		}
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func readInFile(i string) []string {
	// Read line-by-line
	var lines []string
	file, err := os.Open(i)
	if err != nil {
		log.Println(err)
		return lines
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}
