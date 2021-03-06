<h1 align="center"><img src="https://github.com/BenchmarkXPRT/CloudXPRT/blob/master/CloudXPRT-header.png" alt="CloudXPRT Header" /></h1>
<h4 align="center">
  <i>
    A free cloud-native benchmark designed and developed by the
  <a href="https://www.principledtechnologies.com/benchmarkxprt/">BenchmarkXPRT Development community</a>
   </i>
</h4>

<hr>

- [Introduction](#Introduction)
- [CloudXPRT workloads](#CloudXPRT-workloads)
- [Getting started](#Getting-started)
- [Prerequisites](#Prerequisites)
- [Results and results submission](#Results-and-results-submission)
- [Support](#Support)
- [Licensing and legal information](#Licensing-and-legal-information)
- [Get involved with BenchmarkXPRT](#Get-involved-with-BenchmarkXPRT)

## Introduction
[CloudXPRT](https://www.principledtechnologies.com/benchmarkxprt/cloudxprt/) is a cloud benchmark that can accurately measure the performance of modern, cloud-first applications deployed on modern infrastructure as a service (IaaS) platforms, whether those platforms are paired with on-premises (datacenter), private cloud, or public cloud deployments. Regardless of where clouds reside, applications are increasingly using them in latency-critical, highly available, and high-compute scenarios.

CloudXPRT
* Is relevant to on-prem (datacenter), private, and public cloud deployments
* Runs on top of cloud platform software such as Kubernetes and Docker
* Supports multi-tier workloads
* Reports relevant metrics including both throughput and critical latency for responsiveness-driven applications and maximum throughput for applications dependent on batch processing

----
## CloudXPRT workloads
The CloudXPRT Preview release includes two workloads that can be installed and run independently.

### Web microservices
In the web-tier microservices workload, a simulated user logs in to a web application that does three things: provides a selection of stock options, performs Monte-Carlo simulations with those stocks, and presents the user with options that may be of interest. The workload reports performance in transactions per second, which testers can use to directly compare IaaS stacks and to evaluate whether any given stack is capable of meeting service-level agreement (SLA) thresholds.

#### [Set up and install the Web Microservices workload](web-microservices)

### Data analytics
The machine learning (ML) training workload calculates XGBoost model training time. XGBoost is a gradient-boosting framework  that data scientists often use for ML-based regression and classification problems. The purpose of the workload in the context of CloudXPRT is to evaluate how well an IaaS stack enables XGBoost to speed and optimize model training. The workload reports latency and throughput rates. As with the web-tier microservices workload, testers can use this workload’s metrics to compare IaaS stack performance and to evaluate whether any given stack is capable of meeting SLA thresholds.

#### [Set up and install the Data Analytics workload](data-analytics/README.md)

## Prerequisites
We highly recommended running this benchmark on high end servers. While running, the benchmark will scale to utilize all the cores available. However, for functional testing, your physical node or VM must have at least:
* 16 logical or virtual CPUs
* 8 GB RAM
* 10 GB of available disk space (20 GB for the Data Analytics workload)

Each of these test configurations require both Docker and Kubernetes.
Off-premises tests will require access to an Amazon Web Services (AWS), Google Cloud Platform (GCP), or Microsoft Azure account, depending on the desired test configuration.
An internet connection is required.

## Results and results submission
When the web microservices workload is complete, the benchmark saves the results to CloudXPRT_vXXX_web-microservices/cnbrun/output in CSV format, along with a log file.

When the data analytics workload is complete, the benchmark saves the results to CloudXPRT_vXXX_data-analytics/cnbrun/results.csv in CSV format, generated by the command line –./cnb-analytics_parse-all-results.sh | sed -e 's/\s\+/,/g' > results.csv.e. The log file will appear in the same folder.

To submit results to our page, please follow these [instructions](https://www.principledtechnologies.com/benchmarkxprt/cloudxprt/2020/submit-results.php).

To see results published by the BenchmarkXPRT Development Community, visit[ CloudXPRT results](https://www.principledtechnologies.com/benchmarkxprt/cloudxprt/2020/results) page.

## Support
If you have any questions or comments about CloudXPRT, please feel free to contact a BenchmarkXPRT Development Community representative directly by sending a message to BenchmarkXPRTsupport@PrincipledTechnologies.com.

## Licensing and legal information

For legal and licensing information, please see the following file:

* [LICENSE](https://github.com/BenchmarkXPRT/CloudXPRT/blob/master/LICENSE.txt)

## Get involved with BenchmarkXPRT
CloudXPRT is part of the BenchmarkXPRT suite of performance evaluation tools (the XPRTs), which includes AIXPRT, WebXPRT, CrXPRT, TouchXPRT, HDXPRT, and MobileXPRT. The XPRTs help people get the facts before they buy, use, or evaluate tech products such as servers, desktops, laptops, and tablets.

The XPRTs are developed by the BenchmarkXPRT Development Community, a diverse group that includes over 80 corporations and organizations representing major hardware manufacturers, chip vendors, and tech press leaders. The community provides members with the opportunity to contribute to the process of creating and improving the XPRTs. Community members can do all of the following and more:
* Submit comments, suggestions, questions, and concerns that inform the design of future benchmarks
* See the proposal for new versions of the tools and contribute comments for the final design
* Access and run previews of new benchmarks
* Submit source code for possible inclusion in the benchmarks and examine existing source code

We encourage you to add your voice to the XPRT mix. Participation is open to everyone, so get the details and join the community [here](https://www.principledtechnologies.com/benchmarkxprt/forum/register.php). You can also contact a BenchmarkXPRT Development Community representative directly by sending a message to BenchmarkXPRTsupport@PrincipledTechnologies.com.

To learn more about the BenchmarkXPRT Development Community, view our benchmarks, browse test results, and much more, go to www.BenchmarkXPRT.com.
