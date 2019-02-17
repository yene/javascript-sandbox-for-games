const WebSocket = require('ws');

var world = {
  health: 100,
  weapons: 30,
  enemy: {
    name: "slime",
    health: 100,
  },
  players: [
    {
      index: 0,
      name: "peter " + 0,
      health: 100,
    }
  ]
}
for (let i = 1; i < 5000; i++) {
  world.players.push(
    {
      index: i,
      name: "peter " + i,
      health: 100,
    }
  )
}


ws.on('open', function open() {
  console.time('process');
  ws.send(JSON.stringify({
    type: "world",
    data: world,
  }));
  ws.send(JSON.stringify({
    type: "code",
    code: `
    for (var i = 0; i < world.players.length; i++) {
      world.players[i].health = 0;
    }
    `
    ,
  }));
});

ws.on('message', function incoming(data) {
  var parsing = JSON.parse(data.toString());
  console.log('the time it took to kill players:', world.players.length);
  console.log('including parsing the JSON');
  console.timeEnd('process');
  // console.log(JSON.stringify(world, null, 2));
});
