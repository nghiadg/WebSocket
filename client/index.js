function connectWebSocket() {
  const websocket = new WebSocket("ws://127.0.0.1:8000/ws");

  websocket.onopen = (event) => {
    console.log("Connect websocket success!");
  };
}
