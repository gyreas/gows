import { WebSocketServer } from 'ws';

const PORT = 4440;
const wss = new WebSocketServer({ port: PORT });

console.log(`Listening on ${PORT}`)

wss.on("connection", (ws, req) => {
    console.log(`Got connection: ${ws}/${req}`)

    ws.addEventListener("message", (event) => {
       console.log(`Got a message: ${event.data}`)
    });
    ws.on("close", () => {
        console.log("closing")
    });
});
