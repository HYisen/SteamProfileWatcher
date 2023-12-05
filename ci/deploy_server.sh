cd ..
# Assume the server OS may varies from the dev machine. And wish the ABI matches.
env GOOS=linux GOARCH=amd64 go build
# Need ssh config have been setup properly.
scp steamprofilewatcher Bandwagon:~/spw