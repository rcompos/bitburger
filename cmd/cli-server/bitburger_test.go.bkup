package main

import (
	//"github.com/ktrysmt/go-bitbucket"
	"fmt"

	"github.com/rcompos/go-bitbucket"
)

//type RepositoriesOptions struct {
//	Owner string `json:"owner"`
//	Role  string `json:"role"` // role=[owner|admin|contributor|member]
//}

func main() {

	//c := bitbucket.NewBasicAuth("composr", "Ride2day!!")
	c := bitbucket.NewBasicAuth("roncompos", "ride2day")
	//c := bitbucket.NewBasicAuth("rcompos", "Sportster1998")

	opt := &bitbucket.RepositoriesOptions{
		//Owner: "solidfire",
		//Owner: "stumptown",
		//Owner: "orangelo",
		Role:  "admin",
		Teams: "solidfire",
		//RepoSlug: "solidfire",
	}

	//var res *RepositoriesRes
	//res, err = c.Repositories.ListReposForAccount(opt)
	//res, err := c.Repositories.ListForAccount(opt)

	//var t *bitbucket.RepositoriesRes
	//var err error
	//t, err = c.Repositories.ListForAccount(opt)
	//t, err := c.Repositories.ListForAccount(opt)
	t, err := c.Repositories.ListForTeam(opt)
	if err != nil {
		panic(err)
	}

	//fmt.Println("\nt:\n", t)
	fmt.Println("t Page:", t.Page)
	fmt.Println("t Size:", t.Size)
	//fmt.Println("t MaxDepth:", t.MaxDepth)
	fmt.Println("t Pagelen:", t.Pagelen)
	//fmt.Println("\nt Items:\n", t.Items)

	repositories := t.Items

	//fmt.Println("\nrepositories:\n", repositories)

	var repoList []string
	//for k, v := range repositories {
	for _, v := range repositories {
		//fmt.Println("\nk: %v	v: %v\n", k, v)
		//fmt.Printf("> %v  name:  %v\n", k, v.Full_name)
		repoList = append(repoList, v.Full_name)
		//fmt.Printf("> %v\nlinks: %v\n", k, v.Links)
	}

	fmt.Printf("\nRepos:\n")
	for i, j := range repoList {
		fmt.Printf("%v>  %v\n", i, j)
	}

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
