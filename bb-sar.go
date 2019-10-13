package main

import (
	//"github.com/ktrysmt/go-bitbucket"

	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
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

	// Must set BITBUCKET_USERNAME, BITBUCKET_PASSWORD and BITBUCKET_OWNER
	bbUser := os.Getenv("BITBUCKET_USERNAME")
	bbPassword := os.Getenv("BITBUCKET_PASSWORD")
	bbOwner := os.Getenv("BITBUCKET_OWNER")
	bbRole := os.Getenv("BITBUCKET_ROLE")
	//bbReposlug := os.Getenv("BITBUCKET_REPOSLUG")
	//fmt.Println("BITBUCKET_USERNAME:", bbUser)
	//fmt.Println("BITBUCKET_PASSWORD:", bbPassword)
	//fmt.Println("BITBUCKET_OWNER:", bbOwner)
	//fmt.Println("BITBUCKET_REPOSLUG:", bbReposlug)

	// TODO:  Check exists BITBUCKET_USERNAME, BITBUCKET_PASSWORD and BITBUCKET_OWNER

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
	emoji.Printf("Acquiring repos for %s [ git clone :hamburger: | git pull :fries: ]\n\n", bbOwner)
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
		go func(i int, j string, dir string) {
			defer wg.Done()
			//errNum := gitClone(j, dir)
			//gitCloneString := Sprintf("git clone %s/%s", BBURL, j)
			gitClone := "git clone " + BBURL + "/" + j
			errCloneNum := repoAction(j, gitClone, dir, ":hamburger:", "", "", "")
			if errCloneNum == 128 {
				//gitPull(j)
				gitPull := "git pull"
				errPullNum := repoAction(j, gitPull, dir, "", ":fries:", ":poop:", ":thumbsdown:")
				if errPullNum != 0 {
					emoji.Printf(":poop:")
				}
			}
			//
			//gitSearch(j)
			//
			//fmt.Printf("%vth goroutine done.\n", i)
		}(i, j, reposBaseDir+"/"+bbOwner)
	}

	wg.Wait()
	fmt.Printf("\n\nAll goroutines complete.")
	fmt.Println(" Goodbye world!")

} //

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
