document.onreadystatechange = function () {
    clearLog();
    document.getElementById("file-upload-form").reset()
}

const downloadURL = (data, fileName) => {
    const a = document.createElement("a");
    a.href = data;
    a.download = fileName;
    document.body.appendChild(a);
    a.style.display = "none";
    a.click();
    a.remove();
};

const appendLog = (line) => {
    logBox.value += line;
}

const clearLog = () => {
    const logBox = document.getElementById("logBox");
    logBox.value = "";
}

const downloadBlob = (data, fileName, mimeType) => {
    const blob = new Blob([data], {
        type: mimeType,
    });

    const url = window.URL.createObjectURL(blob);
    downloadURL(url, fileName);
    setTimeout(() => window.URL.revokeObjectURL(url), 1000);
};

const worker = new Worker("/js/worker.js");

document
    .getElementById("file-upload-form")
    .addEventListener("submit", function (evt) {
        evt.preventDefault();
        const process_btn = document.getElementById("process-btn");
        process_btn.innerText = "⚙️Processing...";
        process_btn.disabled = true;

        const file = document.getElementById("file-upload-input").files[0];
        clearLog();
        appendLog("[+] Please stand by...\n")

        worker.onmessage = (evt) => {
            const message = evt.data;
            switch (message["type"]) {
                case "file": {
                    const outFile = message["value"];
                    if (outFile.length == 0) {
                        appendLog("[!] Extraction failed");
                    }
                    else {
                        appendLog("[+] Extraction completed successfully, downloading zip");
                        downloadBlob(outFile, file.name + "_extracted.zip", "application/octet-stream");
                    }
                    process_btn.innerText = "⚙️Process";
                    process_btn.disabled = false;
                    break;
                }
                case "log": {
                    appendLog(message["value"]);
                    break;
                }
            }
        };
        worker.postMessage(file);
    });
