faas-cli delete -f post.yml
sleep 5
faas-cli build -f post.yml
faas-cli push -f post.yml
sleep 25
faas-cli deploy -f post.yml
