# Scripts

The scripts contained in this directory facilitate **building** and **testing** 
the plugin.  The main entry point for most of these scripts is the 
[Makefile](../Makefile). 

## Running tests

### small tests

From the root of the project run `make test-small`

### medium tests

From the root of the project run `make test-medium`

### Large

From the root of the projet run `make test-large`

Large tests require that `docker` is available and configured.  By default the 
tests will also require and use `docker-compose` to orchestrate the containers 
used. As an alternative to `docker-compose` kubernetes can used to orchestrate 
the containers by running the test with the env `TEST_K8S=1 make test-large`.

![large-test](http://i.giphy.com/3oz8xNjcjCaqepanvy.gif)