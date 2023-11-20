# VPN setup

This document outlines VPN (or upstream) setup. It's the shortest document in the setup series. It assumes that the server has been already provisioned.

## Overview

Server should have an `upstream` interface created on startup, but it's impossible to use it unless it's set up by WireJump management tool – `wjcli`. For additional security, none of your account information or connection preferences are saved to disk (apart from keys and addressses in interface files), so if the server is rebooted, you will have to perform setup procedure again.

Login to your server and run `wjcli`. It should display all available commands. For example, to display connection status at any time, run `wjcli status`. That's what you get if you didn't setup a VPN provider yet (or issued `wjcli reset`):

```
$ wjcli status
Upstream connection
  Online:               no                           
  Active since:         N/A 
  Country:              N/A                       
  City:                 N/A                     
Provider
  Name:                 N/A                       
  Preferred location:   N/A                       
  Account expires:      N/A 
```

## Quick connect

Let's setup a connection to Mullvad VPN, for example:

```
$ wjcli setup
Incomplete options provided, forcing interactive mode
Provider : mullvad <-- provider name
Username : 1234    <-- your account number without spaces and hyphens
Password :         <-- not needed for Mullvad (just press Enter)
Command executed successfully
```

If you run `wjcli status` now, you can notice the changes:
```
$ wjcli status
Upstream connection
  Online:               no                           <-- not connected yet                           
  Active since:         N/A
  Country:              N/A                      
  City:                 N/A                     
Provider
  Name:                 mullvad                      <-- selected provider                
  Preferred location:   N/A                       
  Account expires:      Sat Mar 10 2001 19:59:59 UTC <-- account validity time appeared 
```

Note, that connection is NOT established yet. To do so, issue `wjcli connect`:
```
$ wjcli connect
Command executed successfully
```

This will connect to a random exit node in a random country:
```
$ wjcli status
Upstream connection
  Online:               yes                            <-- connected!            
  Active since:         Fri Mar 10 2000 19:59:59 UTC   <-- when did connection occur
  Country:              France                         <-- randomly selected location                 
  City:                 Paris                          <-- random exit node in that location            
Provider
  Name:                 mullvad                       
  Preferred location:   N/A                       
  Account expires:      Sat Mar 10 2001 19:59:59 UTC    
```


`upstream` interface should become active, and, if you have set up client correctly, your home network is browsing Internet from Paris now!

## Advanced connect

If you require a specific exit node, first query all available locations:

```
$ wjcli servers
Servers:                Germany, Denmark, UK, Norway, Sweden, Switzerland, Finland, France, Netherlands 
Last updated:           Fri, 10 Mar 2000 19:59:59 UTC                                                                
Preferred location:     N/A    
```

Then, set location preference via `wjcli servers -p` (you can also use interactive mode here with `-i`):
```
$ wjcli servers -p UK
Command executed successfully

$ wjcli status
...
Provider                              
  Preferred location:   UK   <--- This exit location will be used now
...
```

Run `wjcli connect` again to reconnect:
```
$ wjcli connect
Command executed successfully

$ wjcli status
Upstream connection
  Online:               yes                           
  Active since:         Fri Mar 10 2000 19:59:59 UTC
  Country:              UK      <-- matches preferred location                 
  City:                 London  <-- new exit node in that location            
Provider
  Name:                 mullvad                       
  Preferred location:   UK      <-- location preference is saved now                     
  Account expires:      Sat Mar 10 2001 19:59:59 UTC       
```

## Congrats!

Well done! Setup is now complete.
