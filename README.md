
---

# Real-Time Location Streaming Server

This project implements a real-time location streaming server using Go (Golang) with the fasthttp library and WebSocket for communication. It allows clients to connect and receive real-time updates of their location on a map.

## Features

- **Real-Time Updates**: Clients can receive real-time updates of their location on a map.
- **WebSocket Communication**: Uses WebSocket for efficient bidirectional communication between clients and the server.
- **Leaflet.js Integration**: Displays the real-time location on an interactive map using Leaflet.js.
- **Flexible and Scalable**: Built with Go for high performance and scalability.

## Prerequisites

- Go (Golang) installed on your system.
- Basic understanding of WebSocket communication.

## Installation

1. Clone the repository:

    ```bash
    git clone https://github.com/nihankhan/locastream.git
    ```

2. Navigate to the project directory:

    ```bash
    cd locastream
    ```

3. Build and run the server:

    ```bash
    go build
    ./locastream
    ```

4. Access the server in your web browser at `http://localhost:8080/home`.

## Usage

1. Open your web browser and navigate to `http://localhost:8080/home`.
2. You should see a real-time map with your location marker.
3. Connect to the WebSocket server to receive real-time location updates.

## Configuration

- You can configure the server address and port in the `main()` function of `main.go`.
- Adjust WebSocket endpoint or route in the router configuration in `router.go`.

## Dependencies

- [fasthttp](https://github.com/valyala/fasthttp): Fast HTTP package for Go.
- [websocket](https://github.com/fasthttp/websocket): WebSocket implementation for fasthttp.
- [Leaflet.js](https://leafletjs.com/): JavaScript library for interactive maps.

## Contributing

Contributions are welcome! Feel free to open an issue or submit a pull request for any improvements or features you'd like to add.

## License

This project is licensed under the [MIT License](LICENSE).

---
