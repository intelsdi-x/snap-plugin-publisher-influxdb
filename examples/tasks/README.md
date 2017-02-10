# Running an example

## Requirements 
 * `docker` and `docker-compose` are **installed** and **configured** 
 * this plugin [downloaded and configured](../../README.md#installation) 

## Example
[This](psutil-influxdb-http.yml) example task will publish metrics to **influxdb** from a psutil collector plugin.

### Start your container
In the root of the plugin's repository run,
`$ DEMO=true ./scripts/large.sh`

If you want to run this example using udp instead of http, run the following command instead: `$ DEMO=true TASK="psutil-influxdb-udp.yml" ./scripts/large.sh`

Open another terminal and run one of the following commands based on what you want to experiment with:

- **Snap container**: `$ docker exec -it $(docker ps | sed -n 's/\(\)\s*intelsdi\/snap.*/\1/p') /bin/bash`
- **Influxdb container**: `$ docker exec -it $(docker ps | sed -n 's/\(\)\s*tutum\/influxdb.*/\1/p') /bin/bash`

type `exit-program` in your first terminal at any time to quit and shut down the containers. 

![influx_publisher_setup_new2](https://cloud.githubusercontent.com/assets/21182867/22810489/f61158d0-eeed-11e6-87e3-16d62637b2b4.gif)

### Example tasks
**Snap container**: 

In the Snap container you can see loaded plugins, loaded tasks, and available metrics. You can even tap directly into the data stream that Snap is collecting by watching a task. Learn about other commands you can run in your Snap container [here](https://github.com/intelsdi-x/snap/blob/master/docs/SNAPTEL.md), or in our main [Snap directory](https://github.com/intelsdi-x/snap/).

![influx_publisher_snap3](https://cloud.githubusercontent.com/assets/21182867/22764401/644a4ade-ee1f-11e6-82e6-31c929f20393.gif)


**Influxdb container**:

You can access the Influx container directly through the command line or in a web browser at localhost:8083. The task is publishing the collected data directly to influxdb where you can query the database as you normally would. 

The example demonstrated below has `isMultiFields` set to true. This and [other parameters](../../README.md#documentation) can be adjusted in the [task](psutil-influxdb-http.yml). 

Examples of how to enter the container and query the database can be seen in the screen casts below.  

![influx_publisher_docker_terminal2](https://cloud.githubusercontent.com/assets/21182867/22796509/6dbf7568-eeaf-11e6-9e28-5dbb29138651.gif)

![influx_publisher_docker_website](https://cloud.githubusercontent.com/assets/21182867/22764717/38878e0a-ee21-11e6-9686-8491d6f0ee6f.gif)



## How it works:
- running `$ DEMO=true ./scripts/large.sh` spins up the environment we run large tests in and pauses after loading the first task. 
- [This](psutil-influxdb-http.yml) is the default task and is found in the /examples/tasks folder of the plugin.
- When you specify a different task, it looks in /examples/tasks for that file. We offer [psutil-influxdb-udp.yml](psutil-influxdb-udp.yml) as another task option.
- [docker-compose.yml](../../scripts/test/docker-compose.yml) is our docker compose file which defines two linked containers
    - "snap" is the container where snapteld is run from. 
    - "influxdb" is the container running influxdb. 
