+++
author = "Sriram Panyam"
title = "Basic Cost Estimation"
date = "2020-01-21"
description = "Understanding cost are important to highlight a system's performance across several metrics."
tags = [
    "markdown",
    "text",
]
+++

Systems read, transform and write data between components in the pursuit of business goals.  Sounds simple!   The choice of components however affect the overall cost of a system across a number of metrics.   These metrics could be span performance (eg latencies, throughputs), health (error rates and availability), or even costs (plain $$s).   Understanding the performance characteristics of the system (and its components) is thus paramount.

# System's work-centers

A system (eg a VM) or its components spend time in two essential activities:

* **I/O** - Reading or writing data from/between media.  This could be media such disks, memory, or even to remote systems across a network.  Note that the "network" is not treated as a special case.   The "network" is just another medium for communicating with other systems.  The cost of I/O across a network is part of the overall cost of a system.
* **Processing** - This is the time/span spent by a system/component transforming data in-between I/O activities.

## Machine/VM Provisioning Costs

Since ultimately physical or (virtual) machines are the ones powering services, the costs associated with the storage, memory and processing power of these machines is important to understand.

Machines can be hosted (in a data-center) or now more popularly within a cloud provider's offering.   The provisioning costs can be modelled as below.

### Data-Center costs

When self-hosting in a data-center servers are hosted in Racks.  The unit of space in a rack is the [Rack-Unit](https://en.wikipedia.org/wiki/Rack_unit) and data-center servers (or their enclosures) are built to ensure they fit in multiples of rack units.    Enterprise-grade servers are typically built with higher reliability and are typically beefier machines.  Some of the costs are below.  Note that while storage is usually replaceable memory and processing are usually coupled.  Typically higher memory machines have larger number of cores.

| Item | Cost | notes |
| ---- | ---- | ----- |
| Cost of a Rack	| [$45 per Rack Unit per month](https://www.voxility.com/colocation/prices?eqtype=10u)	 | |
| Enterprise HDD	| [$515 for 10Tb](https://www.yobitech.com/Dell-1-2TB-10K-SAS-HDD-s/306.htm)	  | 0.08c per Gb per month over 5 years |
| Enterprise SSD	| [$4300 for 7.8Tb](https://www.yobitech.com/Dell-7-68TB-SSD-SATA-SAS-Drives-s/640.htm)	| 0.1c per GB per month over 5 years |
| Memory	        | [$1600 for 32Gb Dell Server](https://www.dell.com/en-us/work/shop/cty/pdp/spd/poweredge-r730/pe_r730_1356?mkwid=scMK58oqe&pcrid=255034882484&pkw=&pmt=&pdv=c&slid=&product=PE_R730_1356&pgrid=56333245847&ptaid=aud-644024522857:pla-417818532699&ven1=scMK58oqe,255034882484,901qz26673,c,,PE_R730_1356,56333245847,aud-644024522857:pla-417818532699&ven2=,&lid=5957976&dgc=st&dgseg=cbg&acd=12309215337205630&cid=307000&st=&gclid=CjwKCAiAgqDxBRBTEiwA59eEN59QbLrXn4XzRWnE6YaFfsiJMAqCKQSbiyscksR2M6HYm_vFXLdb2xoClWIQAvD_BwE&ven3=110305286736446189&configurationid=a5863a3c-d996-4800-9ae0-0ea4b497c364)	| 0.1c per Gb per hour (over 5 years) |

### Cloud costs

Leveraging the cloud is almost a no-brainer for most businesses mainly due to a lack of core competency (in managing infrastructure) as well as in leveraging economies of scale and scope.   The cloud has matured to offer extremely high level of customizability while still ensuring elasticity and avoiding high fixed costs.

#### Cost of Storage

| Item	| Cost	| Notes |
| ----- | ----- | ----- |
| HDD	    | $0.025 per GB per month	 | |
| SSD	    | $0.1 per GB per month	 | |
| Memory	| $0.01 per 1GB (RAM) per hour = $7.2 per month	| Typically this would depend on the cores but as the memory increases so do number of cores. The beefy m5n.24xlarge instance for example has 96 cores with 384Gb of RAM costing: $5.712 per hour = $4100 per month. |

#### Cost of processing

Processing data is not free.  It requires CPUs (or GPUs).   Processing costs are influenced by the number of cores in a machine (which determines the level of parallelism).  Very often, the number of cores is not completely independent of the amount of memory in a machine.   So the above table has a rough rule-of-thumb on calculating Memory+Processing together.   However in the sytem design exercises and case stadies that follow differnet machine classes (by different cloud vendors) will be considered to play with the costs and performance of a system.

## Data Transfer (I/O) latencies

Data transfer occurs at several levels, between several types of media and across several links (connections).   From Peter Norvig’s famous [table of costs](http://norvig.com/21-days.html#answers):

<table>
  <tbody>
    <tr>
      <td class="has-text-align-center" data-align="center"><strong>Operation</strong></td>
      <td class="has-text-align-center" data-align="center"><strong>Time</strong></td>
    </tr>
    <tr>
      <td class="has-text-align-center" data-align="center">L1 cache reference</td>
      <td class="has-text-align-center" data-align="center">0.5 ns</td>
    </tr>
    <tr>
      <td class="has-text-align-center" data-align="center">Branch mispredicgt</td>
      <td class="has-text-align-center" data-align="center">5 ns</td>
    </tr>
    <tr>
      <td class="has-text-align-center" data-align="center">L2 Cache Reference</td>
      <td class="has-text-align-center" data-align="center">7 ns</td>
    </tr>
    <tr>
      <td class="has-text-align-center" data-align="center">Mutex lock/unlock</td>
      <td class="has-text-align-center" data-align="center">25 ns</td>
    </tr>
    <tr>
      <td class="has-text-align-center" data-align="center">Main memory reference</td>
      <td class="has-text-align-center" data-align="center">100 ns</td>
    </tr>
    <tr>
      <td class="has-text-align-center" data-align="center">Compressing 1k with zippy</td>
      <td class="has-text-align-center" data-align="center">3000 ns</td>
    </tr>
    <tr>
      <td class="has-text-align-center" data-align="center">Send 1k over a 1GBps network</td>
      <td class="has-text-align-center" data-align="center">10,000 ns (10 us)</td>
    </tr>
    <tr>
      <td class="has-text-align-center" data-align="center">Read 4k randomly from SSD (1GBps)</td>
      <td class="has-text-align-center" data-align="center">150,000 ns (150 us)</td>
    </tr>
    <tr>
      <td class="has-text-align-center" data-align="center">Read 1MB sequentially from RAM</td>
      <td class="has-text-align-center" data-align="center">250000 ns (250 us)</td>
    </tr>
    <tr>
      <td class="has-text-align-center" data-align="center">Round trip with a datacenter</td>
      <td class="has-text-align-center" data-align="center">5000000 ns (500 us)</td>
    </tr>
    <tr>
      <td class="has-text-align-center" data-align="center">Read 1MB sequentially from SSD</td>
      <td class="has-text-align-center" data-align="center">1000000 ns (1ms)</td>
    </tr>
    <tr>
      <td class="has-text-align-center" data-align="center">(HDD) Disk Seek</td>
      <td class="has-text-align-center" data-align="center">10,000,000 ns (10ms)</td>
    </tr>
    <tr>
      <td class="has-text-align-center" data-align="center">Read 1MB Sequentially from HDD</td>
      <td class="has-text-align-center" data-align="center">20,000,000 ns (20ms)</td>
    </tr>
    <tr>
      <td class="has-text-align-center" data-align="center">Send a packet from CA -&gt; Netherlands -&gt; CA</td>
      <td class="has-text-align-center" data-align="center">150,000,000 (150ms)</td>
    </tr>
  </tbody>
</table>

Also see a [Humanized Comparison](https://gist.github.com/hellerbarde/2843375)!!

# Key Takeaways

The costs discussed here are at a very basic (almost existential) level in a system's lifecycle.  However systems are complex - especially when the system is not trivial and can composed of several sub systems.  To model the cost of a system there is a need to model the systems enclosed by the parent system.  The costs highlighted allow the system to progressively model costs of systems with increasing complexity.
