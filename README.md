你好！
很冒昧用这样的方式来和你沟通，如有打扰请忽略我的提交哈。我是光年实验室（gnlab.com）的HR，在招Golang开发工程师，我们是一个技术型团队，技术氛围非常好。全职和兼职都可以，不过最好是全职，工作地点杭州。
我们公司是做流量增长的，Golang负责开发SAAS平台的应用，我们做的很多应用是全新的，工作非常有挑战也很有意思，是国内很多大厂的顾问。
如果有兴趣的话加我微信：13515810775  ，也可以访问 https://gnlab.com/，联系客服转发给HR。

<p align="center">
<img
    src="https://user-images.githubusercontent.com/6550035/46709024-9b23ad00-cbf6-11e8-9fb2-ca8b20b7dbec.jpg"
    width="408px" border="0" alt="croc">
<br>
<a href="https://github.com/schollz/croc/releases/latest"><img src="https://img.shields.io/badge/version-4.1.5-brightgreen.svg?style=flat-square" alt="Version"></a>
<img src="https://img.shields.io/badge/coverage-77%25-brightgreen.svg?style=flat-square" alt="Code coverage">
<a href="https://travis-ci.org/schollz/croc"><img
src="https://img.shields.io/travis/schollz/croc.svg?style=flat-square" alt="Build
Status"></a> 
<a href="https://saythanks.io/to/schollz"><img src="https://img.shields.io/badge/Say%20Thanks-!-brightgreen.svg?style=flat-square" alt="Say thanks"></a>
</p>


<p align="center"><code>curl https://getcroc.schollz.com | bash</code></p>

*croc* is a tool that allows any two computers to simply and securely transfer files and folders. There are many tools that can do this but AFAIK *croc* is the only tool that is easily installed and used on any platform, *and* has secure peer-to-peer transferring, *and* has the capability to resume broken transfers. _Note:_ A new WebRTC version is in progress.

## Overview

**Transmit encrypted data with a code phrase**

*croc* securely transfers data using *code phrases* - a combination of three random words (mnemonicoded 4 bytes). The code phrase is shared between the sender and the recipient for password authenticated key exchange ([PAKE](https://github.com/schollz/pake)), a cryptographic method to use a shared weak key (the "code phrase") to generate a strong key for secure end-to-end encryption. By default, a code phrase can only be used once between two parties so an attacker would have a chance of less than 1 in *4 billion* to guess the code phrase correctly to steal the data. An attacker with the wrong code phrase will fail the PAKE and the sender will be notified without any data transfering. Only two people with the right code phrase will be able to computers transfer encrypted data through a relay.

**Fast data transfer through TCP**

The actual data transfer is accomplished using a relay, either using raw TCP sockets or websockets. If both computers are on the LAN network then *croc* will use a local relay, otherwise a public relay is used. All the data going through the relay is encrypted using the PAKE-generated session key, so the relay can't spy on information passing through it. The data is transferred in blocks, where each block is compressed and encrypted, and the recipient keeps track of blocks received so that it can resume the transfer if interrupted.

**Relay allows any two computers to connect**

*croc* differs from a utility like *scp* because it doesn't require any two computers to have enabled port-forwarding. Instead, *croc* will uses a relay - a temporary server setup locally (if both computers are on lan) or publicly (default is at croc4.schollz.com). Any two computers can connect to the relay, and after securing their channel with PAKE, they can transfer encrypted metadata and data through the relay. The relay works by first having the computers communicate the PAKE protocol via websockets, and then exchanging encrypted metadata, and then stapling the TCP connections directly so that they can transfer directly.

**Why another data transfer utility?**

My motivation to write *croc*, as stupid as it sounds, is because I wanted to create a program that made it easy to send a 3GB+ PBS documentary to my friend in a different country. My friend has a Windows computer and is not comfortable using a terminal. So I wanted to write a program that, while secure, is simple to receive a file. *croc* accomplishes this, and now I find myself using it almost everyday at work. To receive a file you can just download the executable and double click on it. The name is inspired by the [fable of the frog and the crocodile](https://web.archive.org/web/20180926035731/http://allaboutfrogs.org/stories/crocodile.html).

## Examples

The first example shows the basic transfer of some file or folder from computer 1 to computer 2. _These two gifs should run in sync if you force-reload (Ctl+F5)_

![send](.github/1.gif)

![receive](.github/2.gif)

The second example shows how you can restart a broken transfer. Here, computer 2 presses Ctl+C during a transfer to abruptly break the connection, and then resumes by having computer 1 re-send the file. _These two gifs should run in sync if you force-reload (Ctl+F5)_

![send](.github/3.gif)

![receive](.github/4.gif)

In version 4, there is now a GUI available for [Windows](https://github.com/schollz/croc/releases/latest) and [Linux](https://github.com/schollz/croc/releases/latest):

<div style="text-align:center">
    <img src="https://user-images.githubusercontent.com/6550035/47256575-8a193e00-d437-11e8-96fa-42c9d072a8f1.PNG">
</div>

## Install

[Download the latest release for your system](https://github.com/schollz/croc/releases/latest).

Or, you can [install Go](https://golang.org/dl/) and build from source with `go get github.com/schollz/croc`. Since *croc* uses [Go modules](https://golang.org/doc/go1.11#modules) it requires Go version 1.11+.

Or, you can quickly install a release from the command-line:

```
$ curl https://getcroc.schollz.com | bash
```

```
$ wget -qO- https://getcroc.schollz.com | bash
```


## Usage 

The basic usage is to just do 

```
$ croc send FILE
```

to send and then on the other computer you can just do 

```
$ croc [code phrase]
```

to receive (you'll be prompted to enter the code phrase). Note, by default, you don't need any arguments for receiving, instead you will be prompted to enter the code phrase. This makes it possible for you to just double click the executable to run (nice for those of us that aren't computer wizards).

### Custom code phrase

You can send with your own code phrase (must be more than 4 characters).

```
$ croc send --code [code phrase] [filename]
```

### Use locally

*croc* automatically will attempt to start a local connection on your LAN to transfer the file much faster. It uses [peer discovery](https://github.com/schollz/peerdiscovery), basically broadcasting a message on the local subnet to see if another *croc* user wants to receive the file. *croc* will utilize the first incoming connection from either the local network or the public relay and follow through with PAKE.

You can change this behavior by forcing *croc* to use only local connections (`--local`) or force to use the public relay only (`--no-local`):

```
$ croc --local/--no-local send [filename]
```

### Using pipes - stdin and stdout

You can easily use *croc* in pipes when you need to send data through stdin or get data from stdout. To send you can just use pipes:

```
$ cat [filename] | croc send
```

In this case *croc* will automatically use the stdin data and send and assign a filename like "croc-stdin-123456789". To receive to stdout at you can always just use the `-yes` and `-stdout` flags which will automatically approve the transfer and pipe it out to stdout. 

```
$ croc -yes -stdout [code phrase] > out
```

All of the other text printed to the console is going to `stderr` so it will not interfere with the message going to stdout.

### Self-host relay

The relay is needed to staple the parallel incoming and outgoing connections. The relay temporarily stores connection information and the encrypted meta information. The default uses a public relay at, `croc4.schollz.com`. You can also run your own relay, it is very easy, just run:

```
$ croc relay
```

Make sure to open up TCP ports (see `croc relay --help` for which ports to open). Relays can also be customized to which elliptic curve they will use (default is siec).

You can send files using your relay by entering `-addr` to change the relay that you are using if you want to custom host your own.

```
$ croc -addr "myrelay.example.com" send [filename]
```

### Configuration file 

You can also make some paramters static by using a configuration file. To get started with the config file just do 

```
$ croc config
```

which will generate the file that you can edit. 
Any changes you make to the configuration file will be applied *before* the command-line flags, if any.


## License

MIT

## Acknowledgements

*croc* has been through many iterations, and I am awed by all the great contributions! If you feel like contributing, in any way, by all means you can send an Issue, a PR, ask a question, or tweet me ([@yakczar](http://ctt.ec/Rq054)).

Thanks...

- ...[@warner](https://github.com/warner) for the [idea](https://github.com/warner/magic-wormhole).
- ...[@tscholl2](https://github.com/tscholl2) for the [encryption gists](https://gist.github.com/tscholl2/dc7dc15dc132ea70a98e8542fefffa28).
- ...[@skorokithakis](https://github.com/skorokithakis) for [code on proxying two connections](https://www.stavros.io/posts/proxying-two-connections-go/).
- ...for making pull requests [@Girbons](https://github.com/Girbons), [@techtide](https://github.com/techtide), [@heymatthew](https://github.com/heymatthew), [@Lunsford94](https://github.com/Lunsford94), [@lummie](https://github.com/lummie), [@jesuiscamille](https://github.com/jesuiscamille), [@threefjord](https://github.com/threefjord), [@marcossegovia](https://github.com/marcossegovia), [@csleong98](https://github.com/csleong98), [@afotescu](https://github.com/afotescu), [@callmefever](https://github.com/callmefever), [@El-JojA](https://github.com/El-JojA), [@anatolyyyyyy](https://github.com/anatolyyyyyy), [@goggle](https://github.com/goggle), [@smileboywtu](https://github.com/smileboywtu), [@nicolashardy](https://github.com/nicolashardy)!
