# About "ZAWS"

Zabbix AWS monitoring template
This template is supported over Zabbix 3.0.

# Features

* Automatically registration EC2 instances information by using Zabbix LLD host prototype.
* Automatically registration EC2 CloudWatch metrics information by using Zabbix LLD item prototype.

# Requirements

* Zabbix >= 3.0

# Installation & Setting

## 1. Download

Download zaws command line tool and Zabbix template xml file.

Please get the binary file that is appropriate for your environment architecture.

[binary file url](https://github.com/ike-dai/zaws/releases)

[template xml file url](https://github.com/ike-dai/zaws/tree/master/templates/zaws_zabbix_template.xml)

## 2. Copy to Externalscripts directory

Please copy command line tool file to your zabbix servers externalscripts directory.

for example:

    $ cp zaws-linux-amd64 /usr/lib/zabbix/externalscripts/zaws

## 3. Import zabbix template xml file

[Configuration]->[Templates]->[Import]

Please import "zaws_zabbix_template.xml"

## 4. Register host

[Configuration]->[Hosts]->[Create host]

* Host name: any
* Groups: any
* Agent interfaces: any (not used in this tool)
* Templates: Template AWS
* Macros: please set 3 macro
    * {$REGION}: Please set AWS region name (e.g. ap-northeast-1)
    * {$KEY}: Please set AWS ACCESS KEY ID (e.g. AKI........)
    * {$SECRET}: Please set AWS SECRET ACCESS KEY

# Contact

Please send feedback to me.

Daisuke IKEDA

Twitter: [@ike_dai](https://twitter.com/ike_dai)

e-mail: <dai.ikd123@gmail.com>

# License

Licensed under the Apache License, Version 2.0. The Apache v2 full text is published at this [link](http://www.apache.org/licenses/LICENSE-2.0).

Copyright 2016 Daisuke IKEDA.
