package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/fatih/color"
	"github.com/kyokomi/emoji"
)

var wg sync.WaitGroup

func main() {

	//var bbUser, bbPassword, bbProject, bbRole, searchStr, replaceStr, inFile, outFile, branch, prTitle string
	var bbUser, bbPassword, bbProject, searchStr, replaceStr, inFile, outFile, branch, prTitle string
	var execute, createPR, gitClone, help bool
	//var numWorkers int

	flag.StringVar(&bbUser, "u", os.Getenv("BBS_USERNAME"), "Bitbucket user (required) (envvar BBS_USERNAME)")
	flag.StringVar(&bbPassword, "p", os.Getenv("BBS_PASSWORD"), "Bitbucket password (required) (envvar BBS_PASSWORD)")
	flag.StringVar(&bbProject, "j", os.Getenv("BBS_PROJECT"), "Bitbucket project (required) (envvar BBS_PROJECT)")
	//flag.StringVar(&bbRole, "e", os.Getenv("BBS_ROLE"), "Bitbucket role (envvar BBS_ROLE)")
	flag.StringVar(&searchStr, "s", os.Getenv("BBS_SEARCH"), "Text to search for (envvar BBS_SEARCH)")
	flag.StringVar(&replaceStr, "r", os.Getenv("BBS_REPLACE"), "Text to replace with (envvar BBS_REPLACE)")
	flag.StringVar(&branch, "b", os.Getenv("BBS_BRANCH"), "Feature branch where changes are made (envvar BBS_BRANCH)")
	flag.StringVar(&prTitle, "t", os.Getenv("BBS_TITLE"), "Title for pull request (envvar BBS_PRTITLE)")
	flag.StringVar(&inFile, "i", "./repos.txt", "Input file of repos (repo project) one per line")
	flag.StringVar(&outFile, "f", "./repos.txt", "Output file")
	flag.BoolVar(&execute, "x", false, "Execute text replace")
	flag.BoolVar(&createPR, "c", false, "Create pull request")
	flag.BoolVar(&gitClone, "g", false, "Git clone repos")
	flag.BoolVar(&help, "h", false, "Help")
	flag.Parse()

	//bbURL := "https://bitbucket.ngage.netapp.com/rest/api/1.0/"
	//bbAPIURL := "https://" + bbUser + ":" + bbPassword + "@bitbucket.ngage.netapp.com/rest/api/1.0/"
	bbURL := "https://" + bbUser + ":" + bbPassword + "@bitbucket.ngage.netapp.com/scm/"
	fmt.Printf("bbURL: %v\n", bbURL)
	var repoCache []string

	if help == true {
		color.Set(color.FgYellow)
		fmt.Printf("BitBurger Server - BitBucket Server Search and Replace\n\n")
		color.Unset()
		color.Set(color.FgMagenta)
		emoji.Printf("[ clone :hamburger: | pull :fries: | changes :cherries: | pull request :fire: ]\n\n")
		color.Unset()
		flag.PrintDefaults()
		os.Exit(1)
	}

	if bbUser == "" || bbPassword == "" || bbProject == "" {
		//if bbUser == "" || bbPassword == "" {
		fmt.Println("Must supply user (-u) and password (-p) and project(-j)!")
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
		promptRead(bbProject, searchStr, replaceStr)
	}

	if createPR && (prTitle == "" || branch == "") {
		fmt.Println("Must supply pull request title (-t) and feature branch name (-b)!")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var repoList []string
	repoMap := make(map[string]string)
	if _, err := os.Stat(inFile); err == nil && inFile != "" {
		fmt.Printf("Using repo input file: %v\n", inFile)
		// inFile exists, use it
		//fmt.Printf("inFile found! %v\n", inFile)
		readDiskCache(&repoCache, inFile)
		for _, j := range repoCache {
			//fmt.Printf(">>  i: %v   j: %v\n", i, j)
			rSlice := strings.Split(j, " ")
			//repoList = append(repoList, j)
			repoMap[rSlice[0]] = rSlice[1]
		}
	} else {
		//  If inFile not exist, then request from BB API

		pjRepoCmd := "curl -s -u " + bbUser + ":" + bbPassword + " https://bitbucket.ngage.netapp.com/rest/api/1.0/projects/" + bbProject + "/repos/ | jq -r '.values[].slug'"
		color.Set(color.FgYellow)
		fmt.Printf("> %v\n", pjRepoCmd)
		color.Unset()
		pjRepoExec := exec.Command("bash", "-c", pjRepoCmd)
		pjRepoExecOut, _ := pjRepoExec.Output()
		pjResult := string(pjRepoExecOut)
		fmt.Printf("pjResult:\n'%v'\n", pjResult)
		repoList = strings.Split(pjResult, "\n")
		for _, rl := range repoList {
			if rl != "" {
				repoMap[rl] = bbProject
				fmt.Printf(">>> '%v' = '%v'\n", rl, bbProject)
			}
		}

		// TODO: Save just repo
		writeDiskCache(repoMap, outFile)

	}

	if !gitClone && !execute && !createPR && searchStr == "" {
		// List repos and exit
		for r := range repoMap {
			fmt.Println(r)
		}
		os.Exit(0)
	}

	color.Set(color.FgMagenta)
	//emoji.Printf("Acquiring repos for %s [ clone :hamburger: | pull :fries: | untracked :cherries: | pull request :fire: ]\n\n", bbProject)
	emoji.Printf("[ clone :hamburger: | pull :fries: | untracked :cherries: | pull request :fire: ]\n\n")
	color.Unset()

	reposBaseDir := "repos"
	scm := "git" // dummy scm
	createDir(fmt.Sprintf("%s/%s", reposBaseDir, bbProject))

	//fmt.Printf("Repos:\n")
	// TODO: Change from waitgroup to buffered channels
	for repo, project := range repoMap {
		//fmt.Printf("%v> %v\n", repo, project)
		if strings.HasPrefix(repo, "#") {
			fmt.Printf("Skipping: %v\n", repo)
			continue
		}
		wg.Add(1)
		repoProject := project + "/" + repo
		go bitBurger(createPR, execute, repoProject, scm, reposBaseDir, project, searchStr, replaceStr, bbUser, bbPassword, branch, prTitle, bbURL)

	}

	wg.Wait()
	//fmt.Printf("\n\nAll goroutines complete.")
	fmt.Println()

} // End main

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

func bitBurger(createPR, execute bool, j, scm, dir, owner, search, replace, user, pw, fBranch, pr, url string) {

	defer wg.Done()
	dirOwner := dir + "/" + owner
	dirRepo := dir + "/" + j
	//fmt.Printf("%s/%s\n", url, j)

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

				/*
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
				*/
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
