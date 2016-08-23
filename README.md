# HiveKind
A proof of concept for starting and managing local and remote processes hierarchically using go. Think htop for node and golang projects.

## What?
Upon running Daemon one is greeted with a nice nCurses-eqse clui. Use up and down arrows to navigate the tree, and enter to open/close entries.

While this is a work in progress, each entry is supposed to represent an instance of a running process, locally or remotely. There is (as of writing this) no way to initialize a process, and the initial client.js is automatically spawned at runtime for example purposes.

Hopefully this will do something useful in the forseeable future. If not then enjoy this screencap.

<img src="https://raw.githubusercontent.com/CreativeCactus/HiveKind/master/record.gif" alt="cap" style="height:250px; width:250px; right: 0px; position:absolute;"></img>

