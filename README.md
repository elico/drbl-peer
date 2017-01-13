drbl-peer
==========
GoLang library for testing domains and ip addresses against a set of dns or dnsrbl or httpbl service.


Links for public DNS blacklists services
-----
 - [Norton dns blacklists information](https://dns.norton.com/faq.html)
 - [OpenDNS Home Internet Security](https://www.opendns.com/home-internet-security/)

Example for squid external_acl_type usage configurations
-----
```
external_acl_type dnsbl_check ipv4 concurrency=200 ttl=15 %DST %SRC %METHOD /opt/bin/squid-external-acl-helper -peers-filename=/opt/bin/peersfile.txt
acl dnsbl_check_acl external dnsbl_check
deny_info http://ngtech.co.il/block_page/?url=%u&domain=%H dnsbl_check_acl

http_access deny dnsbl_check_acl
```

License
-------
Copyright (c) 2016, Eliezer Croitoru
All rights reserved.

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
