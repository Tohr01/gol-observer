const BASE_URL = "127.0.0.1:8888";//external-api-url-here.com
const API_KEY = "your-api-key-here";

let activeID;
let ws;
let logContainer;
document.addEventListener('DOMContentLoaded', () => {
    const sidebar = document.getElementById("sidebar");
    logContainer = document.getElementById("logContainer");
    fetchLogFiles().then(logFiles => {
        this.logFiles = logFiles;
        if (Object.keys(logFiles).length == 0) {
            return
        }
        for (const wsurl in logFiles) {
            sidebar.appendChild(getNewButtonNode(logFiles[wsurl], wsurl))
        }
        activeID = Object.keys(logFiles)[0]
        const firstLogButton = document.getElementById(activeID);
        firstLogButton.classList.add("sidebar-button-active");
        listenToLogFile(firstLogButton);
    });
});

async function fetchLogFiles() {
    try {
        const logFiles = await fetch("https://" + BASE_URL + "/available-logs?api_key=" + API_KEY);
        return await logFiles.json();
    } catch (error) {
        console.error('Error with async/await:', error);
    }
}

function getNewButtonNode(name, wsurl) {
    // Create a new button element
    const button = document.createElement('button');
    button.className = 'sidebar-button';
    button.setAttribute("id", wsurl);
    button.setAttribute("onclick", "listenToLogFile(this)");

    const p1 = document.createElement('p');
    p1.textContent = '#';

    const p2 = document.createElement('p');
    p2.textContent = name;

    button.appendChild(p1);
    button.appendChild(p2);

    return button;
}

// Handling of ws messages
function listenToLogFile(element) {
    const wsurl = element.id;
    if (ws !== undefined) {
        ws.close();
    }
    clearLogContainer();
    const oldActiveButton = document.getElementById(activeID);
    oldActiveButton.classList.remove("sidebar-button-active");
    element.classList.add("sidebar-button-active");
    activeID = wsurl;
    ws = new WebSocket("ws://" + BASE_URL + "/ws/" + wsurl + "?api_key=" + API_KEY);
    ws.addEventListener("message", (event) => {
        const newP = document.createElement("p");
        newP.innerHTML = event.data;
        logContainer.appendChild(newP);
    });
}

function clearLogContainer() {
    logContainer.textContent = "";
}