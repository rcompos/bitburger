package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/kyokomi/emoji"
	"github.com/rcompos/go-bitbucket"
)

const sleepytime time.Duration = 1

var wg sync.WaitGroup

func main() {

	rand.Seed(time.Now().UnixNano())

	var bbUser, bbPassword, bbOwner, bbRole, searchStr, replaceStr, inFile, outFile, branch, prTitle string
	var execute, createPR, gitClone, help bool
	//var numWorkers int

	flag.StringVar(&bbUser, "u", os.Getenv("BITBUCKET_USERNAME"), "Bitbucket user (required) (envvar BITBUCKET_USERNAME)")
	flag.StringVar(&bbPassword, "p", os.Getenv("BITBUCKET_PASSWORD"), "Bitbucket password (required) (envvar BITBUCKET_PASSWORD)")
	flag.StringVar(&bbOwner, "o", os.Getenv("BITBUCKET_OWNER"), "Bitbucket owner (required) (envvar BITBUCKET_OWNER)")
	flag.StringVar(&bbRole, "e", os.Getenv("BITBUCKET_ROLE"), "Bitbucket role (envvar BITBUCKET_ROLE)")
	flag.StringVar(&searchStr, "s", os.Getenv("BITBUCKET_SEARCH"), "Text to search for (envvar BITBUCKET_SEARCH)")
	flag.StringVar(&replaceStr, "r", os.Getenv("BITBUCKET_REPLACE"), "Text to replace with (envvar BITBUCKET_REPLACE)")
	flag.StringVar(&branch, "b", os.Getenv("BITBUCKET_BRANCH"), "Feature branch where changes are made (envvar BITBUCKET_BRANCH)")
	flag.StringVar(&prTitle, "t", os.Getenv("BITBUCKET_TITLE"), "Title for pull request (envvar BITBUCKET_PRTITLE)")
	flag.StringVar(&inFile, "i", "./repos.txt", "Input file of repos (owner/repo) one per line")
	flag.StringVar(&outFile, "f", "./repos.txt", "Output file")
	flag.BoolVar(&execute, "x", false, "Execute text replace")
	flag.BoolVar(&createPR, "c", false, "Create pull request")
	flag.BoolVar(&gitClone, "g", false, "Git clone repos")
	flag.BoolVar(&help, "h", false, "Help")
	//flag.IntVar(&numWorkers, "w", 100, "Number of worker threads")
	flag.Parse()

	bbURL := "https://" + bbUser + ":" + bbPassword + "@bitbucket.org"
	var repoCache []string

	if help == true {
		color.Set(color.FgYellow)
		fmt.Printf("BitBucket Cloud Search and Replace\n\n")
		color.Unset()
		color.Set(color.FgMagenta)
		emoji.Printf("[ clone :hamburger: | pull :fries: | changes :cherries: | pull request :fire: ]\n\n")
		color.Unset()
		flag.PrintDefaults()
		os.Exit(1)
	}

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

	if execute || createPR {
		promptRead(bbOwner, searchStr, replaceStr)
	}

	if createPR && (prTitle == "" || branch == "") {
		fmt.Println("Must supply pull request title (-t) and feature branch name (-b)!")
		flag.PrintDefaults()
		os.Exit(1)
	}

	myName := fmt.Sprintf(path.Base(os.Args[0]))
	logsDir := "logs"
	createDir(logsDir)
	logFileDateFormat := "2006-01-02-150405"
	logStamp := time.Now().Format(logFileDateFormat)
	logfile := logsDir + "/" + myName + "-" + string(logStamp) + ".log"

	logf, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		log.Fatal(err)
	}
	defer logf.Close()
	log.SetOutput(logf) //log.Println("Test log message")

	log.Printf("Current Unix Time: %v\n", time.Now().Unix())

	optRs := &bitbucket.RepositoriesOptions{}
	optRs.Owner = bbOwner
	optRs.Role = bbRole // [owner|admin|contributor|member]

	c := bitbucket.NewBasicAuth(bbUser, bbPassword)

	var repos *bitbucket.RepositoriesRes

	var repoList []string
	repoMap := make(map[string]string)

	optR := &bitbucket.RepositoryOptions{}
	optR.Owner = bbOwner
	var reepo *bitbucket.Repository

	if _, err := os.Stat(inFile); err == nil && inFile != "" {
		fmt.Printf("Using repo input file: %v\n", inFile)
		// inFile exists, use it
		//fmt.Printf("inFile found! %v\n", inFile)
		readDiskCache(&repoCache, inFile)
		//reps := []string{}
		//reps = repoCache
		//fmt.Printf("Repos:\n")
		for _, j := range repoCache {
			//fmt.Printf(">>  i: %v   j: %v\n", i, j)
			rSlice := strings.Split(j, " ")
			//repoList = append(repoList, j)
			repoMap[rSlice[0]] = rSlice[1]
		}
	} else {
		//  If inFile not exist, then request from BB API
		fmt.Printf("Requesting repos from BB Cloud API\n")
		repos, err = c.Repositories.ListForAccount(optRs)
		if err != nil {
			panic(err)
		}
		repositories := repos.Items
		for _, v := range repositories {
			repoList = append(repoList, v.Full_name)
			// // //
			optR.RepoSlug = v.Slug
			reepo, err = c.Repositories.Repository.Get(optR)
			//fmt.Printf("reepo: %v %v %v\n", reepo.Full_name, reepo.Scm, reepo.Slug)
			repoMap[reepo.Full_name] = reepo.Scm
			// // //
			writeDiskCache(repoMap, outFile)

		}
	}

	if !gitClone && !execute && !createPR && searchStr == "" {
		// List repos and exit
		for r := range repoMap {
			fmt.Println(r)
		}
		os.Exit(0)
	}

	color.Set(color.FgMagenta)
	emoji.Printf("Acquiring repos for %s [ clone :hamburger: | pull :fries: | untracked :cherries: | pull request :fire: ]\n\n", bbOwner)
	color.Unset()

	reposBaseDir := "repos"
	createDir(fmt.Sprintf("%s/%s", reposBaseDir, bbOwner))

	//fmt.Printf("Repos:\n")
	// TODO: Change from waitgroup to buffered channels
	for j, scm := range repoMap {
		//fmt.Printf("%v> %v\n", j, scm)
		if strings.HasPrefix(j, "#") {
			fmt.Printf("Skipping: %v\n", j)
			continue
		}
		//time.Sleep(sleepytime) // slow down to avoid api ban
		wg.Add(1)
		go bitBurger(createPR, execute, j, scm, reposBaseDir, bbOwner, searchStr, replaceStr, bbUser, bbPassword, branch, prTitle, bbURL)
	}

	wg.Wait()
	//fmt.Printf("\n\nAll goroutines complete.")
	fmt.Println()

} //

func bitBurger(createPR, execute bool, j, scm, dir, owner, search, replace, user, pw, fBranch, pr, url string) {
	//go func(i int, j string, dir string, owner string, search string, replace string, createPR bool, user string, pw string) {
	defer wg.Done()
	dirOwner := dir + "/" + owner
	dirRepo := dir + "/" + j
	fmt.Printf("%s/%s\n", url, j)

	if scm == "git" {

		gitClone := "git clone " + url + "/" + j
		errCloneNum := doIt(gitClone, dirOwner, ":hamburger:", "", "", "")
		if errCloneNum == 128 {

			gitPullOrigin := "git pull origin `git symbolic-ref refs/remotes/origin/HEAD | sed 's@^refs/remotes/origin/@@' master`"
			errPullOriginNum := doIt(gitPullOrigin, dirRepo, "", "", "", "")
			if errPullOriginNum != 0 {
				emoji.Printf(":poop:")
			}

			//gitPull := "git pull"
			gitPull := "git branch --set-upstream-to=origin/" + fBranch + " " + fBranch
			errPullNum := doIt(gitPull, dirRepo, ":fries:", "", "", "")
			if errPullNum == 1 {
				upstreamCmd := "git push --set-upstream origin " + fBranch
				color.Set(color.FgYellow)
				fmt.Printf("> %v\n", upstreamCmd)
				color.Unset()
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
		color.Set(color.FgYellow)
		fmt.Printf("> %v\n", branchCmd)
		color.Unset()
		branchExec := exec.Command("bash", "-c", branchCmd)
		branchExec.Dir = dirRepo
		branchExecOut, _ := branchExec.Output()
		bResult := string(branchExecOut)
		fmt.Printf(bResult)

		upstreamCmd := "git push --set-upstream origin " + fBranch
		color.Set(color.FgYellow)
		fmt.Printf("> %v\n", upstreamCmd)
		color.Unset()
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
			sarExec := exec.Command("bash", "-c", sar)
			color.Set(color.FgYellow)
			fmt.Printf("> %v\n", sar)
			color.Unset()
			sarExec.Dir = dirRepo
			sarExecOut, err := sarExec.Output()
			if err != nil {
				panic(err)
				fmt.Printf("ERROR: %v\n", err)
			}
			searchResult := string(sarExecOut)
			color.Set(color.FgBlue)
			fmt.Printf(searchResult)
			color.Unset()
		}

		//fmt.Println("dirRepo: ", dirRepo)
		// Check for untracked changes
		gitDiffIndex := "git diff-index --quiet HEAD --;"
		//gitDiffIndex := `git status -s | wc -l | perl -pe's/^\s+(\d+)\s*/$1/'`
		errGitDiffIndex := doIt(gitDiffIndex, dirRepo, "", "", "", "")
		fmt.Printf("errGitDiffIndex: %v\n", errGitDiffIndex)

		if errGitDiffIndex != 0 {
			// Git untracked changes exist
			emoji.Printf(":cherries:")

			// create Pull Request
			if createPR == true {
				commitCmd := "git commit -am'Replace " + search + " with " + replace + "'"
				color.Set(color.FgYellow)
				fmt.Printf("> %v\n", commitCmd)
				color.Unset()
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
				color.Set(color.FgYellow)
				fmt.Printf("> %v\n", pushCmd)
				color.Unset()
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

				titlePR := pr + " [" + path.Base(j) + "]"

				curlPR := fmt.Sprintf("curl -v https://api.bitbucket.org/2.0/repositories/%s/pullrequests "+
					"-u %s:%s --request POST --header 'Content-Type: application/json' "+
					"--data '{\"title\": \"%s\", \"source\": { \"branch\": { \"name\": \"%s\" } } }'", j, user, pw, titlePR, fBranch)

				//fmt.Printf("curlPR:\n%s\n", curlPR)
				color.Set(color.FgYellow)
				fmt.Printf("> %v\n", curlPR)
				color.Unset()
				prExec := exec.Command("bash", "-c", curlPR)
				prExec.Dir = dirRepo
				prExecOut, err := prExec.Output()
				prResult := string(prExecOut)
				if err == nil {
					if !strings.Contains(prResult, "There are no changes to be pulled") {
						emoji.Printf(":fire:")
					}
				}
				if err != nil {
					fmt.Printf("ERROR: %v\n", err)
				}
				fmt.Printf("PR> %v\n", prResult)
			}

		}
	} else {
		fmt.Printf("ERROR: Unsupported SCM: %s %s\n", j, scm)
		log.Printf("ERROR: Unsupported SCM: %s %s\n", j, scm)
	}

	// end git
	//fmt.Printf("%vth goroutine done.\n", i)
}

func doIt(cmdstr, rdir, win, any, fail, fcess string) int {

	cmd := exec.Command("bash", "-c", cmdstr)
	cmd.Dir = rdir
	color.Set(color.FgYellow)
	fmt.Printf("%s\n", cmd)
	color.Unset()

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

func writeDiskCache(c map[string]string, cf string) {
	// If the file doesn't exist, create it, or append to the file
	//f, err := os.OpenFile(cf, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	f, err := os.OpenFile(cf, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	for fullname, scm := range c {
		outString := fmt.Sprintf("%v %v\n", fullname, scm)
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
