document.addEventListener("DOMContentLoaded", () => {
    const fileInput = document.getElementById("modell");

    function downloadModel(event) {
        event.preventDefault();
        console.log("Downloading model...");

        fetch("/supply-chain/download", { method: "GET" })
            .then((response) => {
                if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
                return response.blob();
            })
            .then((blob) => {
                const url = window.URL.createObjectURL(blob);
                const a = document.createElement("a");
                a.href = url;
                a.download = "model.py";
                document.body.appendChild(a);
                a.click();
                document.body.removeChild(a);
                window.URL.revokeObjectURL(url);
                console.log("Model downloaded.");
            })
            .catch((error) => {
                console.error("Download error:", error);
            });
    }

    function uploadModel(event) {
        event.preventDefault();

        const file = fileInput?.files[0];
        if (!file) {
            console.error("No file selected for upload.");
            return;
        }

        console.log("Uploading model...");
        const formData = new FormData();
        formData.append("modell", file);

        fetch("/supply-chain/upload", { method: "POST", body: formData })
            .then((response) => {
                if (!response.ok) {
                    throw new Error(`HTTP error! Status: ${response.status}`);
                }
                return response.json();
            })
            .then((data) => {
                console.log(data.message || "Model uploaded successfully.");
            })
            .catch((error) => {
                console.error("Upload error:", error);
            });
    }


    // Button bindings
    const downloadButton = document.querySelector("[data-action='download-mod']");
    const uploadButton = document.querySelector("[data-action='upload-mod']");

    if (downloadButton) {
        downloadButton.addEventListener("click", downloadModel);
    } else {
        console.warn("Download button not found in the DOM.");
    }

    if (uploadButton) {
        uploadButton.addEventListener("click", uploadModel);
    } else {
        console.warn("Upload button not found in the DOM.");
    }
});
