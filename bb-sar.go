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
//	Role  string `json:"role"` // role=[owner|admin|contributor|member]
//}
const BBURL = "https://bitbucket.org"

func main() {

	rand.Seed(time.Now().UnixNano())

	// Location under present working directory where repos are cloned
	repodir := "repos"

	bbUser := os.Getenv("BITBUCKET_USERNAME")
	bbPassword := os.Getenv("BITBUCKET_PASSWORD")
	//bbOwner := os.Getenv("BITBUCKET_OWNER")
	//bbReposlug := os.Getenv("BITBUCKET_REPOSLUG")

	//fmt.Println("BITBUCKET_USERNAME:", bbUser)
	//fmt.Println("BITBUCKET_PASSWORD:", bbPassword)
	//fmt.Println("BITBUCKET_OWNER:", bbOwner)
	//fmt.Println("BITBUCKET_REPOSLUG:", bbReposlug)

	c := bitbucket.NewBasicAuth(bbUser, bbPassword)

	opt := &bitbucket.RepositoriesOptions{
		//Owner: "solidfire",
		Owner: "stumptown",
		//Owner: "orangelo",
		//Role: "member",
		//Role: "admin",
	}

	/*
		optRepository := &bitbucket.RepositoryOptions{
			Owner: "solidfire",
			//Owner: "stumptown",
			//Owner: "orangelo",
			RepoSlug: "kubespray-and-pray",
		}
	*/

	//res, err := c.Repositories.ListForAccount(opt)

	//var t *bitbucket.RepositoriesRes
	//var err error
	//t, err = c.Repositories.ListForAccount(opt)
	t, err := c.Repositories.ListForAccount(opt)
	//t, err := c.Repositories.ListForTeam(opt)
	if err != nil {
		panic(err)
	}

	//fmt.Println("\nt:\n", t)
	//fmt.Println("t Page:", t.Page)
	//fmt.Println("t Size:", t.Size)
	//fmt.Println("t MaxDepth:", t.MaxDepth)
	//fmt.Println("t Pagelen:", t.Pagelen)
	//fmt.Println("\nt Items:\n", t.Items)

	repositories := t.Items

	//fmt.Println("\nrepositories:\n", repositories)

	createRepoDir(repodir)

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

		fmt.Printf("%v> %v\n", i, j)

		go func(i int, j string) {
			defer wg.Done() // <2>
			gitClone(j)
			fmt.Printf("%vth goroutine sleeping...\n", i)
			time.Sleep(2)
		}(i, j)
	}

	wg.Wait() // <3>
	fmt.Println("All goroutines complete.")

	//fmt.Println("repolist: ", repoList[0])

	fmt.Println("Goodbye world!")

	// Try to get Repository methods to run, such as Get
	/*
		//repoGot, err := c.Repositories.ListForAccount(optRepositories)
		//repoGot, err := c.Repository.Get(optRepository)
		repoGot, err := c.Repository.ListForks(optRepository)
		if err != nil {
			panic(err)
		}

		fmt.Println()
		fmt.Printf("repoGot:\n %v\n", repoGot)
	*/

	/*
		//repos := t.(map[string]interface{})
		repos := t.(map[string]interface{})
		fmt.Printf("Repos:\n %v\n\bn", repos)

		fmt.Printf("Page: %v\n", repos["page"])
		fmt.Printf("Size: %v\n", repos["size"])
		fmt.Printf("Pagelen: %v\n", repos["pagelen"])
		fmt.Printf("Values:\n %v\n", repos["values"])

		//valString := repos["values"]
		valString := repos["values"].([]interface{})
		fmt.Printf("\nvalString: %v\n", valString)
		for k, v := range valString {
			fmt.Printf("\n->  valString %v: %v\n", k, v)
			fmt.Printf("\nName: %v\n", v.(map[string]interface{})["name"])
			fmt.Printf("Links: %v\n", v.(map[string]interface{})["links"])
		}
	*/

} //

func gitClone(r string) {
	//
	// Git clone repository r
	//

	gitCloneString := fmt.Sprintf("git clone %s/%s", BBURL, r)
	//gitCloneCmd := exec.Command("bash", "-c", gitCloneString)
	//gitCloneCmd.Dir = "repos"
	cmd := exec.Command("bash", "-c", gitCloneString)
	cmd.Dir = "repos"
	fmt.Println(cmd)
	//cmd := exec.Command("ls", "-lah")
	//err := gitCloneCmd.Run()

	var waitStatus syscall.WaitStatus
	if err := cmd.Run(); err != nil {
		if err != nil {
			//os.Stderr.WriteString(fmt.Sprintf("Error: %s\n", err.Error()))
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			//errorNumber := waitStatus.ExitStatus()
			////fmt.Printf("Output: %s\n", []byte(fmt.Sprintf("%d", waitStatus.ExitStatus())))
			//fmt.Printf("Output: %s\n", []byte(fmt.Sprintf("%d", errorNumber)))
		}
	} else {
		// Success
		waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
		fmt.Printf("Output: %s\n", []byte(fmt.Sprintf("%d", waitStatus.ExitStatus())))
	}

}

func createRepoDir(repodir string) {
	//
	// Create repos directory repodir
	//
	//	if [ -d "repos" ]; then echo true; else echo false; fi
	mkdirCmd := fmt.Sprintf("if [ ! -d %s ]; then mkdir -m775 %s; fi", repodir, repodir)
	mkdirExec := exec.Command("bash", "-c", mkdirCmd)
	mkdirExecOut, err := mkdirExec.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println(mkdirCmd)
	fmt.Println(string(mkdirExecOut))

}
