package main

import (
	//"github.com/ktrysmt/go-bitbucket"

	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/rcompos/go-bitbucket"
)

//type RepositoriesOptions struct {
//	Owner string `json:"owner"`
//}
const BBURL = "https://bitbucket.org"

func main() {

	rand.Seed(time.Now().UnixNano())

	bbUser := os.Getenv("BITBUCKET_USERNAME")
	bbPassword := os.Getenv("BITBUCKET_PASSWORD")
	bbOwner := os.Getenv("BITBUCKET_OWNER")
	bbRole := os.Getenv("BITBUCKET_ROLE")
	//bbReposlug := os.Getenv("BITBUCKET_REPOSLUG")

	//fmt.Println("BITBUCKET_USERNAME:", bbUser)
	//fmt.Println("BITBUCKET_PASSWORD:", bbPassword)
	//fmt.Println("BITBUCKET_OWNER:", bbOwner)
	//fmt.Println("BITBUCKET_REPOSLUG:", bbReposlug)

	reposBaseDir := "repos"

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

	createRepoDir(fmt.Sprintf("%s/%s", reposBaseDir, bbOwner))

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
		wg.Add(1) // <1>
		//fmt.Printf("%v> %v\n", i, j)
		go func(i int, j string, rdir string) {
			defer wg.Done() // <2>
			errNum := gitClone(j, rdir)
			if errNum == 128 {
				gitPull(j)
			}
			fmt.Printf("%vth goroutine done.\n", i)
			//time.Sleep(2)
		}(i, j, reposBaseDir+"/"+bbOwner)
	}

	wg.Wait() // <3>
	fmt.Println("All goroutines complete.")

	fmt.Println("Goodbye world!")

} //

func gitClone(r string, rdir string) int {
	// Git clone repository r

	gitCloneString := fmt.Sprintf("git clone %s/%s", BBURL, r)
	//gitCloneCmd := exec.Command("bash", "-c", gitCloneString)
	//gitCloneCmd.Dir = "repos"
	cmd := exec.Command("bash", "-c", gitCloneString)
	cmd.Dir = rdir
	fmt.Println(cmd)
	//cmd := exec.Command("ls", "-lah")
	//err := gitCloneCmd.Run()

	var errorNumber int = 0
	var waitStatus syscall.WaitStatus
	if err := cmd.Run(); err != nil {
		if err != nil {
			//os.Stderr.WriteString(fmt.Sprintf("Error: %s\n", err.Error()))
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			errorNumber = waitStatus.ExitStatus()
			////fmt.Printf("Output: %s\n", []byte(fmt.Sprintf("%d", waitStatus.ExitStatus())))
			//fmt.Printf("Output: %s\n", []byte(fmt.Sprintf("%d", errorNumber)))
		}
	} else {
		// Success
		waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
		fmt.Printf("Output: %s\n", []byte(fmt.Sprintf("%d", waitStatus.ExitStatus())))
	}
	return errorNumber

}

func gitPull(r string) {
	fmt.Println("git pull %s", r)
}

/*
	// Git pull repository r

	gitPullString := fmt.Sprintf("git pull")
	//gitCloneCmd := exec.Command("bash", "-c", gitPullString)
	//gitCloneCmd.Dir = "repos"
	cmd := exec.Command("bash", "-c", gitPullString)
	cmd.Dir = "repos"
	fmt.Println(cmd)
	//cmd := exec.Command("ls", "-lah")
	//err := gitCloneCmd.Run()

	var errorNumber int = 0
	var waitStatus syscall.WaitStatus
	if err := cmd.Run(); err != nil {
		if err != nil {
			//os.Stderr.WriteString(fmt.Sprintf("Error: %s\n", err.Error()))
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			errorNumber = waitStatus.ExitStatus()
			////fmt.Printf("Output: %s\n", []byte(fmt.Sprintf("%d", waitStatus.ExitStatus())))
			//fmt.Printf("Output: %s\n", []byte(fmt.Sprintf("%d", errorNumber)))
		}
	} else {
		// Success
		waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
		fmt.Printf("Output: %s\n", []byte(fmt.Sprintf("%d", waitStatus.ExitStatus())))
	}
	return errorNumber

}
*/

func createRepoDir(repodir string) {
	//
	// Create repos directory repodir
	//
	//	if [ -d "repos" ]; then echo true; else echo false; fi
	mkdirCmd := fmt.Sprintf("if [ ! -d %s ]; then mkdir -p -m775 %s; fi", repodir, repodir)
	mkdirExec := exec.Command("bash", "-c", mkdirCmd)
	mkdirExecOut, err := mkdirExec.Output()
	if err != nil {
		panic(err)
	}
	//fmt.Println(mkdirCmd)
	fmt.Printf(string(mkdirExecOut))

}
