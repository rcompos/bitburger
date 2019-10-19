# bb-sar
BitBucket Cloud Search and Replace.

[ 🍔  | 🍟  | 💎 | 🔥 ]

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
$ go run bb-sar.go -h
BitBucket Cloud Search and Replace

[ clone 🍔  | pull 🍟  | untracked 💎  | pull request 🔥  ]

  -b string
    	Feature branch where changes are made (envvar BITBUCKET_BRANCH)
  -c	Create pull request
  -e string
    	Bitbucket role (envvar BITBUCKET_ROLE)
  -f string
    	Output file (default "./logs/out.txt")
  -g	Git clone repos
  -h	Help
  -i string
    	Input file of repos (owner/repo) one per line
  -o string
    	Bitbucket owner (required) (envvar BITBUCKET_OWNER)
  -p string
    	Bitbucket password (required) (envvar BITBUCKET_PASSWORD)
  -r string
    	Text to replace with (envvar BITBUCKET_REPLACE)
  -s string
    	Text to search for (envvar BITBUCKET_SEARCH)
  -t string
    	Title for pull request (envvar BITBUCKET_PRTITLE)
  -u string
    	Bitbucket user (required) (envvar BITBUCKET_USERNAME)
  -x	Execute text replace
exit status 1
```


Set credentials with CLI arguments or environmental variables.

```
export BITBUCKET_USERNAME=<username>
export BITBUCKET_PASSWORD=<password>
```

### Examples

List all repos under BitBucket Cloud owner \<owner\>.

```
$ go run bb-sar.go -u <user> -p <password> -o <owner>
```

For all repos under BitBucket Cloud owner \<owner\>, search for 'docker.example.net' and replace with 'artifactory.example.net', create feature branch 'HCI-5165-docker-to-artifactory' and create pull request with title 'HCI-5165 :fire: Artifactory Docker Registry DNS docker.example.net -> artifactory.example.net'.  Credentials are set with envvars.

```
$ go run bb-sar.go -x -s 'docker.example.net' -r 'artifactory.example.net' -c -b 'HCI-5165-dns-docker-to-artifactory' -t 'HCI-5165 :fire: Artifactory Docker Registry DNS docker.example.net -> artifactory.example.net'
```

Same thing, but repos read from input file, one per line.

```
$ go run bb-sar.go -i tmp/myrepos.txt -x -s 'docker.example.net' -r 'artifactory.example.net' -c -b 'HCI-5165-dns-docker-to-artifactory' -t 'HCI-5165 :fire: Artifactory Docker Registry DNS docker.example.net -> artifactory.example.net'
```

Clone all repos under BitBucket Cloud owner \<owner\>.

```
$ go run bb-sar.go -u <user> -p <password> -o <owner> -g
```

Search all repos under BitBucket Cloud owner \<owner\> for \<term\>.

```
$ go run bb-sar.go -u <user> -p <password> -o <owner> -s <term>
```
