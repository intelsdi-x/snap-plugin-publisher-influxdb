# Example tasks

[This](task-meminfo-influxdb.yml) example task will publish metrics to **influxdb** 
from a meminfo collector plugin.  

## Running the example

### Requirements 
 * `docker` and `docker-compose` are **installed** and **configured** 

Running the sample is as *easy* as running the script `./run-meminfo-influxdb.sh`. 

![example01](http://i.giphy.com/l2Sq8p7Wyg2rlI2J2.gif)

Note: If you want to run the example without going through Docker you could 
update the task manifest ([task-meminfo-influxdb.yml](task-meminfo-influxdb.yml)) to 
point to your instance of Influxdb using the correct username/password pair and 
then run `meminfo-influxdb.sh`.  

## Files

- [run-meminfo-influxdb.sh](run-meminfo-influxdb.sh) 
    - The example is launched with this script     
- [task-meminfo-influxdb.yml](task-meminfo-influxdb.yml)
    - Snap task definition
- [docker-compose.yml](docker-compose.yml)
    - A docker compose file which defines two linked containers
        - "runner" is the container where snapteld is run from.  You will be dumped 
        into a shell in this container after running 
        [run-meminfo-influxdb.sh](run-meminfo-influxdb.sh).  Exiting the shell will 
        trigger cleaning up the containers used in the example.
        - "influxdb" is the container running influxdb. 
- [meminfo-influxdb.sh](meminfo-influxdb.sh)
    - Downloads `snapteld`, `snaptel`, `snap-plugin-collector-meminfo`,
    `snap-plugin-publisher-influxdb` and starts the task 
    [task-meminfo-influxdb.yml](task-meminfo-influxdb.yml).
- [.setup.sh](.setup.sh)
    - Verifies dependencies and starts the containers.  It's called 
    by [run-meminfo-influxdb.sh](run-meminfo-influxdb.sh).