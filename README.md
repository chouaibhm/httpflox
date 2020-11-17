# httpflox
Probe of list of IPs or domains for working http and https servers and commun web service ports

# Install
``` $> go get -u github.com/chouaibhm/httpflox ```

# Usage
```
$> httpflox -h
Usage of httpflox:
  -c int
        set the concurrency (default 100)
  -method string
        HTTP method to use (default "HEAD")
  -silent
        run in silent mode 
  -t int
        timeout (milliseconds) default 3000 (default 3000)

```
## Run httpflox 
``` $> cat ips.txt | httpflox

http://example.com [200]  [ IIS Windows Server ]
https://192.168.1.54 [200]  [ Welcome]
https://192.168.1.14:9090 [Jboss]
https://example.com:8443 [302] [  ]
```
## Run httpflox in silent mode and Pipe output to file
``` $> cat ips.txt | httpflox | tee output.txt

http://example.com
https://192.168.1.54
https://192.168.1.14:9090
https://example.com:8443
```
