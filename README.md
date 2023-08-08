# CloudRecon
CloudRecon tool

# Install
** You must have CGO enabled, and may have to install gcc to run CloudRecon**
```sh
sudo apt install gcc
```

```sh
go install github.com/g0ldencybersec/CloudRecon@latest
```

# Description
**CloudRecon** 


CloudRecon is a suite of tools for red teamers and bug hunters to find ephemeral and development assets in their campaigns and hunts. 

Often, target organizations stand up cloud infrastructure that is not tied to their ASN or related to known infrastructure. Many times these assets are development sites, IT product portals, etc. Sometimes they don't have domains at all but many still need HTTPs.

CloudRecon is a suite of tools to scan IP addresses or CIDRs (ex: cloud providers IPs) and find these hidden gems for testers, by inspecting those SSL certificates.

The tool suite is three parts in GO:

Scrape - A LIVE running tool to inspect the ranges for a keywork in SSL certs CN and SN fields in real time.

Store - a tool to retrieve IPs certs and download all their Orgs, CNs, and SANs. So you can have your OWN cert.sh database.

Retr - a tool to parse and search through the downloaded certs for keywords.
