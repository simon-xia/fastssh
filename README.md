# FastSSH

FastSSH is a password free tool with a look-up table in local


![](demo.gif)


## Install

download pre-build packages from [releases](https://github.com/simon-xia/fastssh/releases/tag/v1.0), and create a config file:

	echo "name|host|user|password|port|comment\nservice1|192.168.1.1|admin|admin|22| service1 test env" > ~/.fastsshrc


## Config Format

default config file is local at `~/.fastsshrc`, a example is 


	name|host|user|password|port|comment
	service1|192.168.1.1|admin|admin|22| service1 test env
	service2|192.168.1.2|admin|admin|22| service2 stable env

## Usage

default run mode is search, as demo shows.


	fastssh


another mode is run command with host number in your config file


	fastssh <host number>

## Acknowledgments

Thanks [fzf](https://github.com/junegunn/fzf) for their powerful search engine.
