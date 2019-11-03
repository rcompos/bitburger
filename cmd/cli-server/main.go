package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/kyokomi/emoji"
	"github.com/rcompos/bitburger"
)

func main() {

	var wg sync.WaitGroup
	//var bbUser, bbPassword, bbProject, bbRole, searchStr, replaceStr, inFile, outFile, branch, prTitle string
	var bbUser, bbPassword, bbProject, searchStr, replaceStr, inFile, branch, prTitle string
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
	flag.StringVar(&inFile, "i", "", "Input file of project/repo one per line")
	flag.BoolVar(&execute, "x", false, "Execute text replace")
	flag.BoolVar(&createPR, "c", false, "Create pull request")
	flag.BoolVar(&gitClone, "g", false, "Git clone repos")
	flag.BoolVar(&help, "h", false, "Help")
	flag.Parse()

	//bbURL := "https://bitbucket.ngage.netapp.com/rest/api/1.0/"
	//bbAPIURL := "https://" + bbUser + ":" + bbPassword + "@bitbucket.ngage.netapp.com/rest/api/1.0/"
	bbURL := "https://" + bbUser + ":" + bbPassword + "@bitbucket.ngage.netapp.com/scm/"
	bbURLClean := "https://" + bbUser + ":" + "****" + "@bitbucket.ngage.netapp.com/scm/"
	fmt.Printf("bbURL: %v\n", bbURLClean)
	var repoCache []string
	outDir := "repos"
	bitburger.CreateDir(outDir)

	if help == true {
		color.Set(color.FgYellow)
		fmt.Printf("BitBurger Server - BitBucket Server Search and Replace\n\n")
		color.Unset()
		color.Set(color.FgMagenta)
		emoji.Printf("[ clone :fire: | pull :fries: | changes :beer: | pull request :hamburger: ]\n\n")
		color.Unset()
		flag.PrintDefaults()
		os.Exit(1)
	}

	if bbProject != "" && inFile != "" {
		fmt.Println("Specify project (-j) OR infile (-i)")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if bbUser == "" || bbPassword == "" {
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
		if bbProject == "" {
			bbProject = "confirm"
		}
		bitburger.PromptRead(bbProject, searchStr, replaceStr)
	}

	if createPR && (prTitle == "" || branch == "") {
		fmt.Println("Must supply pull request title (-t) and feature branch name (-b)!")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var repoList, slugList []string
	if _, err := os.Stat(inFile); err == nil && inFile != "" {
		fmt.Printf("Using repo input file: %v\n", inFile)
		bitburger.ReadDiskCache(&repoCache, inFile)
		for _, j := range repoCache {
			//fmt.Printf(">>  i: %v   j: %v\n", i, j)
			repoList = append(repoList, j)
		}
	} else {

		if bbProject == "" {
			fmt.Println("Specify project (-j) OR infile (-i)")
			flag.PrintDefaults()
			os.Exit(1)
		}
		//  Default behavior call BB API

		// API Pagination
		// Need to follow API Pagination
		// API Pagination

		pjRepoCmd := "curl -s -u " + bbUser + ":" + bbPassword + " https://bitbucket.ngage.netapp.com/rest/api/1.0/projects/" + bbProject + "/repos/?limit=1000 | jq -r '.values[].slug'"
		pjRepoCmdClean := "curl -s -u " + bbUser + ":" + "****" + " https://bitbucket.ngage.netapp.com/rest/api/1.0/projects/" + bbProject + "/repos/?limit=1000 | jq -r '.values[].slug'"
		color.Set(color.FgYellow)
		fmt.Printf("> %v\n", pjRepoCmdClean)
		color.Unset()
		pjRepoExec := exec.Command("bash", "-c", pjRepoCmd)
		pjRepoExecOut, _ := pjRepoExec.Output()
		pjResultTmp := string(pjRepoExecOut)
		pjResult := strings.TrimSpace(pjResultTmp)
		//fmt.Printf("pjResult:\n'%v'\n\n", pjResult)
		slugList = strings.Split(pjResult, "\n")
		for _, reep := range slugList {
			repoList = append(repoList, bbProject+"/"+reep)
		}
		//fmt.Printf("repoList:\n'%v'\n\n", repoList)
		outFile := outDir + "/" + bbProject + "-repos.txt"
		bitburger.WriteDiskCache(repoList, outFile)
	}

	if !gitClone && !execute && !createPR && searchStr == "" {
		// List repos and exit
		for _, r := range repoList {
			fmt.Println(r)
		}
		os.Exit(0)
	}

	color.Set(color.FgMagenta)
	//emoji.Printf("Acquiring repos for %s [ clone :hamburger: | pull :fries: | untracked :cherries: | pull request :fire: ]\n\n", bbProject)
	emoji.Printf("[ clone :fire: | pull :fries: | untracked :beer: | pull request :hamburger: ]\n\n")
	color.Unset()

	//reposBaseDir := "repos"
	scm := "git" // dummy scm
	//createDir(fmt.Sprintf("%s/%s", reposBaseDir, bbProject))

	//fmt.Printf("Repos:\n")
	// TODO: Change from waitgroup to buffered channels
	for _, repo := range repoList {
		//fmt.Printf("%v> %v\n", repo, project)
		if strings.HasPrefix(repo, "#") {
			fmt.Printf("Skipping: %v\n", repo)
			continue
		}

		reepParts := strings.Split(repo, "/")
		bbProject = strings.ToLower(reepParts[0])
		repoOnly := strings.ToLower(reepParts[1])

		projRepoDir := fmt.Sprintf("%s/%s", outDir, bbProject)
		//fmt.Printf("Create dir: %s\n", projRepoDir)
		bitburger.CreateDir(projRepoDir)

		wg.Add(1)
		//repoProject := bbProject + "/" + repo
		//go bitBurger(createPR, execute, repoProject, scm, reposBaseDir, bbProject, searchStr, replaceStr, bbUser, bbPassword, branch, prTitle, bbURL)
		go bitburger.Sar(createPR, execute, repoOnly, scm, outDir, bbProject, searchStr, replaceStr, bbUser, bbPassword, branch, prTitle, bbURL, wg)

	}

	wg.Wait()
	//fmt.Printf("\n\nAll goroutines complete.")

} // End main
