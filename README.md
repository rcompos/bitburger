# bb-sar
BitBucket search and replace across all repos by owner.

Requires Go v1.13.1 or later

```
$ go run bb-sar.go  -h
Usage of bb-sar:
  -b string
    	Feature branch where changes are made (envvar BITBUCKET_BRANCH)
  -c	Create pull request
  -e string
    	Bitbucket role (envvar BITBUCKET_ROLE)
  -f string
    	Output file (default "./logs/out.txt")
  -i string
    	Input file
  -l	Return repo list only
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
exit status 2
```

### Example usage

For all repos under BitBucket owner \<owner\>, search for 'docker.example.net' and replace with 'artifactory.example.net', create feature branch 'docker-to-artifactory' and create pull request with title 'HCI-5165 Artifactory Docker Registry docker.example.net -> artifactory.example.net'.

```
$ go run bb-sar.go -c -x -u <user> -p <password> -o <owner> -s 'docker.example.net' -r 'artifactory.example.net' -b 'docker-to-artifactory' -t 'HCI-5165 Artifactory Docker Registry docker.example.net -> artifactory.example.net'
```

List all repos under BitBucket owner \<owner\>.

```
$ go run bb-sar.go -l -u <user> -p <password> -o <owner>
```

Clone all repos under BitBucket owner \<owner\>.

```
$ go run bb-sar.go -u <user> -p <password> -o <owner>
```

Search all repos under BitBucket owner \<owner\> for \<term\>.

```
$ go run bb-sar.go -u <user> -p <password> -o <owner> -s <term>
```