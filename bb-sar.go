package main

import (
	//"github.com/ktrysmt/go-bitbucket"

	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/kyokomi/emoji"
	"github.com/rcompos/go-bitbucket"
)

const BBURL = "https://bitbucket.org"

//const sleepytime time.Duration = 2

func main() {

	rand.Seed(time.Now().UnixNano())

	var bbUser, bbPassword, bbOwner, bbRole, searchStr, replaceStr string
	var execute, createPR bool
	//var numWorkers int
	flag.StringVar(&bbUser, "u", os.Getenv("BITBUCKET_USERNAME"), "Bitbucket user (required) (envvar BITBUCKET_USERNAME)")
	flag.StringVar(&bbPassword, "p", os.Getenv("BITBUCKET_PASSWORD"), "Bitbucket password (required) (envvar BITBUCKET_PASSWORD)")
	flag.StringVar(&bbOwner, "o", os.Getenv("BITBUCKET_OWNER"), "Bitbucket owner (required) (envvar BITBUCKET_OWNER)")
	flag.StringVar(&bbRole, "e", os.Getenv("BITBUCKET_ROLE"), "Bitbucket role (envvar BITBUCKET_ROLE)")
	flag.StringVar(&searchStr, "s", os.Getenv("BITBUCKET_SEARCH"), "Text to search for (envvar BITBUCKET_SEARCH)")
	flag.StringVar(&replaceStr, "r", os.Getenv("BITBUCKET_REPLACE"), "Text to replace with (envvar BITBUCKET_REPLACE)")
	flag.BoolVar(&execute, "x", false, "Execute text replace")
	flag.BoolVar(&createPR, "c", false, "Create pull request")
	//flag.IntVar(&numWorkers, "w", 100, "Number of worker threads")
	flag.Parse()

	if bbUser == "" || bbPassword == "" || bbOwner == "" {
		fmt.Println("Must supply user (-u), password (-p) and owner (-o)!")
		fmt.Println("Alternately, environmental variables can be set.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	//searchStr := "docker.artifactory.solidfire.net"
	//replaceStr := "DOCKER.SOLIDFIRE.NET"
	if execute == true && (searchStr == "" || replaceStr == "") {
		fmt.Println("Must supply search string (-s) and replace string (-r)!")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if execute {
		PromptRead(bbOwner, searchStr, replaceStr)
	}

	logsDir := "logs"
	createDir(logsDir)
	// Create log file
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

	color.Set(color.FgMagenta)
	//fmt.Println(path.Base(os.Args[0]))
	emoji.Printf("Acquiring repos for %s [ git clone :hamburger: | git pull :fries: | untracked :gem:]\n\n", bbOwner)
	color.Unset() // Don't forget to unset
	//greeting := ":hamburger:"
	//emoji.Println(strings.Repeat(greeting, 20))
	//time.Sleep(2)

	opt := &bitbucket.RepositoriesOptions{}
	opt.Owner = bbOwner
	opt.Role = bbRole // [owner|admin|contributor|member]

	c := bitbucket.NewBasicAuth(bbUser, bbPassword)

	//repos, err := c.Repositories.ListForAccount(opt)
	repos, err := c.Repositories.ListForAccount(opt)
	if err != nil {
		panic(err)
	}

	//fmt.Println("\nt:\n", repos)
	//fmt.Println("repos Page:", repos.Page)
	//fmt.Println("repos Size:", repos.Size)
	//fmt.Println("repos MaxDepth:", repos.MaxDepth)
	//fmt.Println("repos Pagelen:", repos.Pagelen)
	//fmt.Println("\nrepos Items:\n", repos.Items)

	repositories := repos.Items
	//fmt.Println("\nrepositories:\n", repositories)

	reposBaseDir := "repos"
	createDir(fmt.Sprintf("%s/%s", reposBaseDir, bbOwner))

	var repoList []string
	//for k, v := range repositories {
	for _, v := range repositories {
		//fmt.Println("\nk: %v	v: %v\n", k, v)
		//fmt.Printf("> %v  name:  %v\n", k, v.Full_name)
		repoList = append(repoList, v.Full_name)
		//fmt.Printf("> %v\nlinks: %v\n", k, v.Links)
	}

	var wg sync.WaitGroup

	//fmt.Printf("Repos:\n")
	for i, j := range repoList {
		wg.Add(1)
		//fmt.Printf("%v> %v\n", i, j)
		go func(i int, j string, dir string, owner string, searchStr string, replaceStr string, createPR bool, c *bitbucket.Client) {
			defer wg.Done()
			dirOwner := dir + "/" + owner
			gitClone := "git clone " + BBURL + "/" + j
			errCloneNum := repoAction(j, gitClone, dirOwner, ":hamburger:", "", "", "")
			if errCloneNum == 128 {
				gitPull := "git pull"
				errPullNum := repoAction(j, gitPull, dirOwner, "", ":fries:", ":lemon:", ":lemon:")
				if errPullNum != 0 {
					emoji.Printf(":poop:")
				}
			}

			//# The following one-liner shell command will search and replace all files recursively.
			//# DIR=.; OLD='hello'; NEW='H3110'; find $DIR -type f -print -exec grep -Iq . {} \; -exec perl -pe"s/$OLD/$NEW/g" {} \;

			dirRepo := dir + "/" + j
			fmt.Printf("%s\n", dirRepo)

			var sar string
			if execute == true {
				sar = `find . -path ./.git -prune -o -type f -print  -exec grep -Iq . {} \; -exec perl -i -pe"s/` +
					searchStr + `/` + replaceStr + `/g" {} \;`
			} else if searchStr != "" && replaceStr != "" {
				sar = `find . -path ./.git -prune -o -type f -print -exec grep -Iq . {} \; -exec perl -ne" print if s/` +
					searchStr + `/` + replaceStr + `/g" {} \;`
			} else if searchStr != "" {
				sar = `find . -path ./.git -prune -o -type f -print -exec grep -Iq . {} \; -exec perl -ne" print if /` +
					searchStr + `/g" {} \;`
			}

			if sar != "" {
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

			// Check for untracked changes
			gitDiffIndex := "git diff-index --quiet HEAD --"
			errPullNum := repoAction(j, gitDiffIndex, dirOwner, "", "", "", "")
			if errPullNum != 0 {
				// Git untracked changes exist
				emoji.Printf(":gem:")
				// create Pull Request

				/*
					if createPR == true {
						optPR := &bitbucket.PullRequestsOptions{}
						optPR.Title = "TEST-PULL-REQUEST"

						resultPR, err := c.PullRequests.Create(optPR)
						if err != nil {
							panic(err)
						}
					}
				*/

			}

			//walkDir(dirRepo)

			//fmt.Printf("%vth goroutine done.\n", i)
		}(i, j, reposBaseDir, bbOwner, searchStr, replaceStr, createPR, c)
	}

	wg.Wait()
	fmt.Printf("\n\nAll goroutines complete.")
	fmt.Println(" Goodbye world!")

} //

func walkDir(d string) {

	var subDirToSkip = ".git"
	//err := filepath.Walk(".",
	err := filepath.Walk(d,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && info.Name() == subDirToSkip {
				//fmt.Printf("skipping a dir without errors: %+v \n", info.Name())
				return filepath.SkipDir
			}
			//fmt.Printf("visited file or dir: %q\n", p)
			fmt.Printf("%s\n", p)

			// Perform search
			checkFileInfo(p)

			return nil
		})
	if err != nil {
		log.Println(err)
	}

}

func repoAction(r string, cmdstr string, rdir string, win string, any string, fail string, fcess string) int {

	cmd := exec.Command("bash", "-c", cmdstr)
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

func PromptRead(m string, s string, r string) {
	reader := bufio.NewReader(os.Stdin)
	if s == "" {
		fmt.Printf("Git clone all %s repos.\n", m)
	} else {
		fmt.Printf("Perform text replacement in all %s repos.\n", m)
		fmt.Printf("%s -> %s\n", s, r)
	}
	fmt.Printf("To continue type '%s': \n", m)
	text, _ := reader.ReadString('\n')
	answer := strings.TrimRight(text, "\n")
	//fmt.Printf("answer: %s \n", answer)
	//if answer == "y" || answer == "Y" {
	if answer == m {
		return
	} else {
		//prompt2() //For recursive prompting
		log.Fatal("Exiting without action.")
	}
}
