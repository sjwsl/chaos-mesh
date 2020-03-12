# Network Chaos Document

This document describes how to add network chaos experiments in Chaos Mesh.

Network Chaos actions are mainly divided into the following two categories:

- **Network Partition** action separates pods into several independent subnets by blocking communication between them.

- **Network Emulation (Netem) Chaos** actions cover regular network faults, such as network delay, duplication, loss, and corruption.

## Network Partition Action

Below is a sample network partition configuration file:

```yaml
apiVersion: pingcap.com/v1alpha1
kind: NetworkChaos
metadata:
  name: network-partition-example
  namespace: chaos-testing
spec:
  action: partition
  mode: one
  selector:
    labelSelectors:
      "app.kubernetes.io/component": "tikv"
  direction: to
  target:
    selector:
      namespaces:
        - tidb-cluster-demo
      labelSelectors:
        "app.kubernetes.io/component": "tikv"
    mode: one
  duration: "10s"
  scheduler:
    cron: "@every 15s"
```

> For more sample files, see [examples](../examples). You can edit them as needed. 

Description:

* **action** defines the specific chaos action for the pod. In this case, it is network partition.
* **mode** defines the mode to run chaos action.
* **selector** specifies the target pods for chaos injection.
* **direction** specifies the partition direction. Supported directions are from, to, and both.
* **target** specifies the target for network partition.
* **duration** defines the duration for each chaos experiment. In the sample file above, the network partition lasts for 10 seconds.
* **scheduler** defines the scheduler rules for the running time of the chaos experiment. For more rule information, see <https://godoc.org/github.com/robfig/cron>.


## Netem Chaos Actions

There are 4 cases for netem chaos actions, namely loss, delay, duplicate, and corrupt.

> **Note:** 
> 
> The detailed description of each field in the configuration template are consistent with that in [Network Partition](#network-partition-action).

### Network Loss

A Network Loss action causes network packets to drop randomly. To add a Network Loss action, locate and edit the corresponding template in [/examples](../examples/network-loss-example.yaml).
> In this case, two action specific attributes are required - loss and correlation.
>
> ```yaml
> loss:
>   loss: "25"
>   correlation: "25"
> ```
> **loss** defines the percentage of packet loss.
>
> Network chaos variation isn't purely random, so to emulate that there is a correlation value as well.

### Network Delay

A Network Delay action causes delays in message sending. To add a Network Delay action, locate and edit the corresponding template in [/examples](../examples/network-delay-example.yaml).
> In this case, three action specific attributes are required - correlation, jitter, and latency.
>
>```yaml
>  delay:
>    latency: "90ms"
>    correlation: "25"
>    jitter: "90ms"
>```
> **latency** defines the delay time in sending packets.
>
> **jitter** specifies the jitter of the delay time.
>
> In the above example, the network latency is 90ms ± 90ms.

### Network Duplicate

A Network Duplicate action causes packet duplication. To add a Network Duplicate action, locate and edit the corresponding template in [/examples](../examples/network-duplicate-example.yaml).
> In this case, two attributes are required - correlation and duplicate.
>
>```yaml
>  duplicate:
>    duplicate: "40"
>    correlation: "25"
>```
>
>  **duplicate** indicates the percentage of packet duplication. In the above example, the duplication rate is 40%. 

### Network Corrupt

A Network Corrupt action causes packet corruption. To add a Network Corrupt action, locate and edit the corresponding template in [/examples](../examples/network-corrupt-example.yaml).
> In this case, two action specific attributes are required - correlation and corrupt.
>
>```yaml
>  corrupt:
>    corrupt: "40"
>    correlation: "25"
>```
>
> **corrupt** specifies the percentage of packet corruption.
