package bitburger

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"syscall"

	"github.com/fatih/color"
	"github.com/kyokomi/emoji"
)

//var wg sync.WaitGroup

func HelpMe(msg string) {
	if msg != "" {
		fmt.Printf("%s\n\n", msg)
	}
	color.Set(color.FgYellow)
	fmt.Printf("BitBurger Server - BitBucket Server Search and Replace\n\n")
	color.Unset()
	color.Set(color.FgMagenta)
	emoji.Printf("[ clone :fire: | pull :fries: | changes :beer: | pull request :hamburger: ]\n\n")
	color.Unset()
	flag.PrintDefaults()
	os.Exit(1)
}

func PromptRead(owner string, s string, r string) {
	reader := bufio.NewReader(os.Stdin)
	if s == "" {
		fmt.Printf("Git clone all %s repos.\n", owner)
	} else {
		fmt.Printf("Perform text replacement in all %s repos.\n", owner)
		//fmt.Printf("%s -> %s\n", s, r)
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

func ReadDiskCache(c *[]string, cf string) {
	//fmt.Println("Reading cache from disk: ", cf)
	var lines []string
	lines = readInFile(cf)
	for _, v := range lines {
		(*c) = append(*c, v)
	}
}

func CreateDir(dir string, dbug bool) {
	//	if [ -d "repos" ]; then echo true; else echo false; fi
	mkdirCmd := fmt.Sprintf("if [ ! -d %s ]; then mkdir -p -m775 %s; fi", dir, dir)
	mkdirExec := exec.Command("bash", "-c", mkdirCmd)
	mkdirExecOut, err := mkdirExec.Output()
	check(err)
	if dbug {
		fmt.Println(mkdirCmd)
	}
	fmt.Printf(string(mkdirExecOut))

}

func WriteDiskCache(c []string, cf string) {
	// If the file doesn't exist, create it
	f, err := os.Create(cf)
	check(err)

	for _, value := range c {
		fmt.Fprintln(f, value)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
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

func Sar(createPR, execute, debug bool,
	bbFQDN, repo, dir, ownerIn, search, replace, user, pw, fBranch, pr, url string,
	wg *sync.WaitGroup) {
	// Search-and-replace

	defer wg.Done()

	owner := strings.ToLower(ownerIn)
	j := owner + "/" + repo

	dirOwner := dir + "/" + owner
	fmt.Printf("dirOwner: '%s'\n", dirOwner)
	dirRepo := dir + "/" + j
	fmt.Printf("dirRepo: '%s'\n", dirRepo)

	gitClone := "git clone " + url + "/" + j
	errCloneNum := doIt(gitClone, dirOwner, ":fire:", "", "", "", debug)

	color.Set(color.FgMagenta)
	fmt.Printf("###  %s  ###\n", dirRepo)
	color.Unset()
	fmt.Printf("TEST: %s\n", dirOwner)
	if _, err := os.Stat(dirOwner); err != nil {
		// does not exist
		fmt.Printf("ERROR: Could not clone %s\n", dirRepo)
		return
	}

	if errCloneNum == 128 {

		gitPullOrigin := "git pull origin `git symbolic-ref refs/remotes/origin/HEAD | sed 's@^refs/remotes/origin/@@' master`"
		errPullOriginNum := doIt(gitPullOrigin, dirRepo, "", "", "", "", debug)
		if errPullOriginNum != 0 {
			emoji.Printf(":poop:")
		}

		gitPull := "git branch --set-upstream-to=origin/" + fBranch + " " + fBranch
		errPullNum := doIt(gitPull, dirRepo, ":fries:", "", "", "", debug)
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

	defBranchCmd := "git symbolic-ref refs/remotes/origin/HEAD"
	color.Set(color.FgYellow)
	fmt.Printf("defBranchCmd> %v\n", defBranchCmd)
	color.Unset()
	defBranchExec := exec.Command("bash", "-c", defBranchCmd)
	defBranchExec.Dir = dirRepo
	defBranchExecOut, _ := defBranchExec.Output()
	defBranch := path.Base(strings.TrimSpace(string(defBranchExecOut)))
	fmt.Printf("defBranch> '%s'\n", defBranch)

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
			return
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
	errGitDiffIndex := doIt(gitDiffIndex, dirRepo, "", "", "", "", debug)
	fmt.Printf("errGitDiffIndex: %v\n", errGitDiffIndex)

	if errGitDiffIndex != 0 {
		// Git untracked changes exist
		emoji.Printf(":beer:")

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
			pushResult := string(pushExecOut)
			fmt.Printf(pushResult)

			titlePR := pr + " [" + owner + "/" + repo + "]"
			//curlPR := fmt.Sprintf("curl -v https://api.bitbucket.org/2.0/repositories/%s/pullrequests "+
			//"--data '{\"title\": \"%s\", \"source\": { \"branch\": { \"name\": \"%s\" } } }'", owner, repo, user, pw, titlePR, fBranch)
			curlPR := "curl -v https://" + bbFQDN + "/rest/api/1.0/projects/" + owner + "/repos/" + repo + "/pull-requests " +
				"-u " + user + ":" + pw + " --request POST --header 'Content-Type: application/json' " +
				"--data '{\"title\": \"" + titlePR + "\", " +
				"\"fromRef\": { \"id\": \"refs/heads/" + fBranch + "\", \"repository\": { \"slug\": \"" + repo + "\", \"project\": { \"key\": \"" + owner + "\"} } }," +
				"\"toRef\": { \"id\": \"refs/heads/" + defBranch + "\", \"repository\": { \"slug\": \"" + repo + "\", \"project\": { \"key\": \"" + owner + "\"} } }" +
				"}'"
			//"{ \"name\": \"%s\" } } }'"

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
					emoji.Printf(":hamburger:")
				}
			}
			if err != nil {
				fmt.Printf("ERROR: %v\n", err)
			}
			fmt.Printf("PR> %v\n", prResult)
		}

	}

	//fmt.Printf("%vth goroutine done.\n", i)
	return

}

func doIt(cmdstr, rdir, win, any, fail, fcess string, dbug bool) int {

	cmd := exec.Command("bash", "-c", cmdstr)
	cmd.Dir = rdir
	if dbug {
		color.Set(color.FgYellow)
		fmt.Printf("%s\n", cmd)
		color.Unset()
	}

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
