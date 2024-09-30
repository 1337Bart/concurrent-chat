const socket = new WebSocket('ws://' + window.location.host + '/ws');
let activeRooms = {};
let currentRoom = '';

socket.onopen = function(event) {
    console.log("WebSocket connection established");
};



socket.onerror = function(error) {
    console.error("WebSocket error:", error);
};

socket.onclose = function(event) {
    console.log("WebSocket connection closed:", event);
};

function displayMessage(message) {
    if (!activeRooms[message.room]) {
        activeRooms[message.room] = [];
    }
    activeRooms[message.room].push(message);
    if (currentRoom === message.room) {
        appendMessageToDOM(message);
    }
}

function appendMessageToDOM(message) {
    const messageElement = document.createElement('div');
    messageElement.textContent = `${message.sender}: ${message.content}`;
    document.getElementById('message-container').appendChild(messageElement);
}

document.getElementById('send-btn').addEventListener('click', function() {
    const input = document.getElementById('message-input');
    if (input.value.trim() === '' || !currentRoom) return;
    const message = {
        type: 'chat',
        content: input.value,
        sender: 'User',
        timestamp: new Date(),
        room: currentRoom
    };
    console.log("Sending message:", message);
    socket.send(JSON.stringify(message));
    input.value = '';
});

document.getElementById('create-room-btn').addEventListener('click', function() {
    const roomName = document.getElementById('new-room-input').value;
    if (roomName.trim() === '') return;
    fetch('/room/' + roomName, { method: 'POST' })
        .then(response => {
            if (response.ok) {
                addRoomToList(roomName);
                document.getElementById('new-room-input').value = '';
            }
        });
});

function addRoomToList(roomName) {
    const li = document.createElement('li');
    li.textContent = roomName;
    li.addEventListener('click', function() {
        joinRoom(roomName);
    });
    document.getElementById('rooms').appendChild(li);
}

function joinRoom(roomName) {
    console.log("Joining room:", roomName);
    if (currentRoom !== roomName) {
        socket.send(JSON.stringify({type: 'join', room: roomName}));
        currentRoom = roomName;
        document.getElementById('room-name').textContent = 'Room: ' + roomName;
        document.getElementById('message-container').innerHTML = '';
        if (activeRooms[roomName]) {
            activeRooms[roomName].forEach(appendMessageToDOM);
        }
    }
}

socket.onmessage = function(event) {
    console.log("Received message:", event.data);
    const message = JSON.parse(event.data);
    if (message.type === 'chat') {
        displayMessage(message);
    } else if (message.type === 'join' || message.type === 'leave') {
        console.log(message.type + ' event for room: ' + message.room);
    }
};


fetch('/rooms')
    .then(response => response.json())
    .then(rooms => {
        rooms.forEach(addRoomToList);
    });