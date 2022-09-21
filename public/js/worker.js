importScripts("/js/pyinstxtractor-go.js");

const logFn = (line) => {
    postMessage({
        type: "log",
        value: line
    });
}

onmessage = async (evt) => {
  const file =  evt.data;
  const fileData = new Uint8Array(await file.arrayBuffer());
  const result = extract_exe(file.name, fileData, logFn)

  postMessage({
    type: "file",
    value: result
  });

};
