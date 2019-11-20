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
	var bbFQDN, bbUser, bbPassword, bbProject, searchStr, replaceStr, inFile, branch, prTitle string
	var execute, createPR, gitClone, help, debug bool
	//var numWorkers int

	flag.StringVar(&bbFQDN, "f", os.Getenv("BBS_FQDN"), "BitBucket Server FQDN (required) (envvar BBS_FQDN)")
	flag.StringVar(&bbUser, "u", os.Getenv("BBS_USERNAME"), "Bitbucket user (required) (envvar BBS_USERNAME)")
	flag.StringVar(&bbPassword, "p", os.Getenv("BBS_PASSWORD"), "Bitbucket password (required) (envvar BBS_PASSWORD)")
	flag.StringVar(&bbProject, "j", os.Getenv("BBS_PROJECT"), "Bitbucket project (required) (envvar BBS_PROJECT)")
	flag.StringVar(&searchStr, "s", os.Getenv("BBS_SEARCH"), "Text to search for (envvar BBS_SEARCH)")
	flag.StringVar(&replaceStr, "r", os.Getenv("BBS_REPLACE"), "Text to replace with (envvar BBS_REPLACE)")
	flag.StringVar(&branch, "b", os.Getenv("BBS_BRANCH"), "Feature branch where changes are made (envvar BBS_BRANCH)")
	flag.StringVar(&prTitle, "t", os.Getenv("BBS_TITLE"), "Title for pull request (envvar BBS_PRTITLE)")
	flag.StringVar(&inFile, "i", "", "Input file of project/repo one per line")
	flag.BoolVar(&execute, "x", false, "Execute text replace")
	flag.BoolVar(&createPR, "c", false, "Create pull request")
	flag.BoolVar(&gitClone, "g", false, "Git clone repos")
	flag.BoolVar(&debug, "d", false, "Debugging output")
	flag.BoolVar(&help, "h", false, "Help")
	flag.Parse()

	//bbFQDN := "bitbucket.ngage.netapp.com"
	bbURL := "https://" + bbUser + ":" + bbPassword + "@" + bbFQDN + "/scm/"
	bbURLClean := "https://" + bbUser + ":" + "****" + "@" + bbFQDN + "/scm/"
	var repoCache []string

	if help == true {
		bitburger.HelpMe("")
	}

	//if bbProject != "" && inFile != "" {
	//	bitburger.HelpMe("Specify project (-j) OR infile (-i)")
	//}

	if bbUser == "" || bbPassword == "" {
		bitburger.HelpMe("Must supply user (-u) and password (-p)!\n" +
			"Alternately, environmental variables can be set.")
	}

	if execute == true && (searchStr == "" || replaceStr == "") {
		bitburger.HelpMe("Must supply search string (-s) and replace string (-r)!")
	}

	if execute || createPR {
		if bbProject == "" {
			bbProject = "confirm"
		}
		bitburger.PromptRead(bbProject, searchStr, replaceStr)
	}

	if createPR && (prTitle == "" || branch == "") {
		bitburger.HelpMe("Must supply pull request title (-t) and feature branch name (-b)!")
	}

	if debug {
		fmt.Printf("bbURL: %v\n", bbURLClean)
	}
	outDir := "repos"
	bitburger.CreateDir(outDir, debug)

	curlLimit := "1000"

	var projectList, projects []string

	if bbProject == "" {
		projectListCmd := "curl -s -u " + bbUser + ":" + bbPassword +
			" https://" + bbFQDN + "/rest/api/1.0/projects/?limit=" + curlLimit + " | jq -r '.values[].key'"

		projectListCmdExec := exec.Command("bash", "-c", projectListCmd)
		projectListCmdOut, _ := projectListCmdExec.Output()
		projectListResult := strings.TrimSpace(string(projectListCmdOut))
		projectList = strings.Split(projectListResult, "\n")
		for _, pjName := range projectList {
			projects = append(projects, pjName)
		}
	} else {
		projects = append(projects, bbProject)
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

		for i, pj := range projects {

			if debug {
				fmt.Printf(">>  i: %v   pj: %v\n", i, pj)
			}
			//if bbProject == "" {
			//	fmt.Println("Specify project (-j) OR infile (-i)")
			//	os.Exit(5)
			//}

			// Todo: API Pagination instead of large limit
			pjRepoCheckAuth := "curl -s -u " + bbUser + ":" + bbPassword +
				" https://" + bbFQDN + "/rest/api/1.0/projects/" + pj + "/repos/?limit=10 | jq '.errors'"
			pjRepoCheckAuthClean := "curl -s -u " + bbUser + ":" + "****" +
				" https://" + bbFQDN + "/rest/api/1.0/projects/" + pj + "/repos/?limit=10 | jq '.errors.'"

			if debug {
				color.Set(color.FgYellow)
				fmt.Printf("> %v\n", pjRepoCheckAuthClean)
				color.Unset()
			}
			pjRepoCheckAuthExec := exec.Command("bash", "-c", pjRepoCheckAuth)
			pjRepoCheckAuthOut, _ := pjRepoCheckAuthExec.Output()
			pjRepoCheckAuthResult := strings.TrimSpace(string(pjRepoCheckAuthOut))
			if pjRepoCheckAuthResult != "null" {
				fmt.Printf("%v\n", pjRepoCheckAuthClean)
				fmt.Printf("%v\n", pjRepoCheckAuthResult)
				os.Exit(7)
			}

			//  Default behavior call BB API
			pjRepoCmd := "curl -s -u " + bbUser + ":" + bbPassword + " https://" + bbFQDN + "/rest/api/1.0/projects/" +
				pj + "/repos/?limit=" + curlLimit + " | jq -r '.values[].slug'"
			pjRepoCmdClean := "curl -s -u " + bbUser + ":" + "****" + " https://" + bbFQDN + "/rest/api/1.0/projects/" +
				pj + "/repos/?limit=" + curlLimit + " | jq -r '.values[].slug'"

			if debug {
				color.Set(color.FgYellow)
				fmt.Printf("> %v\n", pjRepoCmdClean)
				color.Unset()
			}
			pjRepoExec := exec.Command("bash", "-c", pjRepoCmd)
			pjRepoExecOut, _ := pjRepoExec.Output()
			pjResult := strings.TrimSpace(string(pjRepoExecOut))
			slugList = strings.Split(pjResult, "\n")
			for _, reep := range slugList {
				repoList = append(repoList, pj+"/"+reep)
			}
			//fmt.Printf("repoList:\n'%v'\n\n", repoList)
			outFile := outDir + "/" + pj + "-repos.txt"
			bitburger.WriteDiskCache(repoList, outFile)
		}
	}

	if !gitClone && !execute && !createPR && searchStr == "" {
		// List repos and exit
		for _, r := range repoList {
			fmt.Println(r)
		}
		os.Exit(0)
	}

	color.Set(color.FgMagenta)
	emoji.Printf("[ clone :fire: | pull :fries: | untracked :beer: | pull request :hamburger: ]\n\n")
	color.Unset()

	for _, repo := range repoList {
		//fmt.Printf("%v> %v\n", repo, project)
		if strings.HasPrefix(repo, "#") {
			fmt.Printf("Skipping: %v\n", repo)
			continue
		}

		reepParts := strings.Split(repo, "/")
		bbPJ := strings.ToLower(reepParts[0])
		repoOnly := strings.ToLower(reepParts[1])
		fmt.Printf("%s %s\n", bbPJ, repoOnly)

		projRepoDir := fmt.Sprintf("%s/%s", outDir, bbPJ)
		bitburger.CreateDir(projRepoDir, debug)

		wg.Add(1)
		go bitburger.Sar(createPR, execute, debug, bbFQDN, repoOnly, outDir, bbPJ, searchStr, replaceStr, bbUser, bbPassword, branch, prTitle, bbURL, &wg)

	}

	wg.Wait()
	fmt.Println()

} // End main
