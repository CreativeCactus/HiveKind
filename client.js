/// A node client with polyglot interfaces. Typically started with a socket file
// sprites stolen from ... http://untamed.wild-refuge.net/rmxpresources.php?characters
var jf = require('jsonfile');
var mtime = require('microtime.js')
var PORT = 3003
var cwd=process.cwd()
var nodes = {}
var MY_ID=Date.now()

/*

    HTTP server
    
*/

var app = require('express')();
var server = require('http').createServer(app);

app.get("/",(req,res)=>{
	res.sendFile(process.cwd()+'/client.html')	
})
app.get("/jq.js",(req,res)=>{
	res.sendFile(process.cwd()+'/jquery-3.0.0.min.js')	
})
app.get("/cytoscape.min.js",(req,res)=>{
	res.sendFile(process.cwd()+'/cytoscape.min.js')	
})

/*

    socket server 
    
*/

var io = require('socket.io')(server);
server.listen(PORT,()=>{ console.log(`http://127.0.0.1:${PORT}/`) });

io.on('connection', function(socket){
    socket.on('hello', function(msg){
        var UID=msg.id
        console.log(`New client: ${UID}`)
        for(var n in nodes)       io.emit('addNode',nodes[n]);
    });
})

/*

    Net client to hivemaster
    
*/

var net = require('net');
var client = net.createConnection("/tmp/hivemaster.sock");

client.on("connect", function() {
	client.write(`hello|{"id":"${MY_ID}"}`)
});

client.on("data", function(data) {
	console.log('>>> '+data)
	process.exit();
});

/*

    Utils

*/

//takes an array of objects and an object to match
function where(list, crit){
    o=list.constructor()
    for(var i in list)if(list.hasOwnProperty(i)){
        var l=list[i], match = true
        for(c in crit)if(crit.hasOwnProperty(c))
            match=match&&(l[c]==crit[c])
        if(match)(o.push?o.push(l):o[i]=l)
    }
    return o
}

//any property defined in crit will be passed back
function mask(i,crit){
    o={}
    for(var c in crit)if(crit.hasOwnProperty(c)) o[c]=i[c]
    return o
}

//index an object or array by the value of a given x property
function varies(o,x){
    out={}
    o.map((v,i,a)=>{
        v._id=i
        var V=v[x]
        if(!out[V]) out[V]=[]
        out[V].push(v)        
    })
    return out
}

//random id
function grid(){
    return ~~(Math.random()*(1<<24))
}