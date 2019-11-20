# BitBurger
BitBurger - BitBucket Cloud Search and Replace.

 [ 🔥 | 🍟 | 🍺 | 🍔 ]

Perform actions for all repos by owner OR a list of repos (owner/repo) from input file.
Default action is to list all repos for owner.

	* List
	* Search
	* Search and Replace
	* Create Pull Requests


### Requires

Go v1.13.1 or later  
git, perl, jq  

### Usage

Make sure your git username and email are configured:

```
$ git config --global user.name "FIRST_NAME LAST_NAME"  
$ git config --global user.email "user@example.com"  
```

```
$ go run ./cmd/cli-server/main.go -h
BitBurger Server - BitBucket Server Search and Replace

[ clone 🔥  | pull 🍟  | changes 🍺  | pull request 🍔  ]

  -b string
    	Feature branch where changes are made (envvar BBS_BRANCH)
  -c	Create pull request
  -g	Git clone repos
  -h	Help
  -i string
    	Input file of project/repo one per line
  -j string
    	Bitbucket project (required) (envvar BBS_PROJECT)
  -p string
    	Bitbucket password (required) (envvar BBS_PASSWORD)
  -r string
    	Text to replace with (envvar BBS_REPLACE)
  -s string
    	Text to search for (envvar BBS_SEARCH)
  -t string
    	Title for pull request (envvar BBS_PRTITLE)
  -u string
    	Bitbucket user (required) (envvar BBS_USERNAME)
  -x	Execute text replace
```


Set credentials with CLI arguments or environmental variables.

```
export BITBUCKET_USERNAME=<username>
export BITBUCKET_PASSWORD=<password>
export BITBUCKET_PROJECT=<owner>
```

### Examples

List all repos under BitBucket Server project \<owner\>.

```
$ go run ./cmd/cli-server/main.go -u <user> -p <password> -j <project>
```

User, password and owner are set with envvars.  For all repos under BitBucket Cloud owner \<owner\>, search for 'docker.example.net' and replace with 'artifactory.example.net', create feature branch 'HCI-5165-docker-to-artifactory' and create pull request with title 'HCI-5165 :fire: Artifactory Docker Registry DNS docker.example.net -> artifactory.example.net'.

```
$ go run ./cmd/cli-server/main.go -x -s 'docker.example.net' -r 'artifactory.example.net' -c -b 'HCI-5165-dns-docker-to-artifactory' -t 'HCI-5165 :fire: Artifactory Docker Registry DNS docker.example.net -> artifactory.example.net'
```

Same thing, but repos read from input file, one per line.

```
$ go run ./cmd/cli-server/main.go -i tmp/myrepos.txt -x -s 'docker.example.net' -r 'artifactory.example.net' -c -b 'HCI-5165-dns-docker-to-artifactory' -t 'HCI-5165 :fire: Artifactory Docker Registry DNS docker.example.net -> artifactory.example.net'
```

Clone all repos under BitBucket Server owner \<owner\>.

```
$ go run ./cmd/cli-server/main.go -u <user> -p <password> -j <project> -g
```

Search all repos under BitBucket Server owner \<project\> for \<term\>.

```
$ go run ./cmd/cli-server/main.go -u <user> -p <password> -j <project> -s <term>
```
