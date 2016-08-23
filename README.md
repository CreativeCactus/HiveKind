# HiveKind
A proof of concept for starting and managing local and remote processes hierarchically using go. Like htop for node and golang projects.

## What?
Upon running Daemon one is greeted with a nice nCurses-eqse clui. Use up and down arrows to navigate the tree, and enter to open/close entries.

While this is a work in progress, each entry is supposed to represent an instance of a running process, locally or remotely. There is (as of writing this) no way to initialize a process, and the initial client.js is automatically spawned at runtime for example purposes.

Hopefully this will do something useful in the forseeable future. If not then enjoy this screencap.

<img src="https://raw.githubusercontent.com/CreativeCactus/HiveKind/master/record.gif" alt="cap" style="height:250px; width:250px; right: 0px; position:absolute;"></img>

There is also a web ui which is in the process of being removed, since the clui is much more attractive. 

## Why?

I wanted to create a way of initializing projects with the boilerplate abstracted away. I often find myself running my ArduinoScope project just to use the REPL, then I paste in my 3-line snippet. It's much faster than setting up my Johnny-five boilerplate just for that. 

It would be nice to create an alias, like 

```
    alias node-j5="node /etc/j5template.js "
    node-j5 mycode.js
``` 

But I would lose track of my project templates quickly, and would prefer to modularize certain aspects of these 'launchers'. For example, it might be a little tedious to write templates to dockerize some go code, compile it and spit out the binary (like https://github.com/remoteinterview/compilebox - I can't find the tool I'm thinking of, but this is similar).
