let inputEditor;
let outputEditor;
let customMapper = ""; // Store the custom mapper from modal

// Load EHR folders on page load and init CodeMirror
window.onload = function () {
    // Initialize CodeMirror for input
    inputEditor = CodeMirror(document.getElementById("inputBox"), {
        mode: "application/json",
        theme: "material-darker",
        lineNumbers: true,
        lineWrapping: true,
        tabSize: 2,
    });

    // Initialize CodeMirror for output
    outputEditor = CodeMirror(document.getElementById("outputBox"), {
        mode: "application/json",
        theme: "material-darker",
        lineNumbers: true,
        lineWrapping: true,
        readOnly: true,
        tabSize: 2,
    });

    // Optionally set default content
    inputEditor.setValue(`[
  {
    "patientid": "123",
    "firstname": "John",
    "lastname": "Doe"
  }
]`);

    // Fetch EHR folder list
    fetch("/folders")
        .then(res => res.json())
        .then(data => {
            const dropdown1 = document.getElementById("dropdown1");
            dropdown1.innerHTML = '<option disabled selected>Select an EHR</option>';
            data.forEach(folder => {
                const option = document.createElement("option");
                option.value = folder;
                option.text = folder;
                dropdown1.appendChild(option);
            });
        });
};

// Load files when EHR folder is selected
function loadFiles() {
    const folder = document.getElementById("dropdown1").value;
    fetch(`/files?folder=${folder}`)
        .then(res => res.json())
        .then(data => {
            const dropdown2 = document.getElementById("dropdown2");
            dropdown2.innerHTML = '<option disabled selected>Select a resourceType</option>';
            data.forEach(file => {
                const option = document.createElement("option");
                option.value = file;
                option.text = file;
                dropdown2.appendChild(option);
            });
        });
}

// Open the custom mapper modal
function openCustomMapperModal() {
    document.getElementById("customMapperModal").style.display = "block";
}

// Store custom mapper from modal
function storeCustomMapper() {
    customMapper = document.getElementById("customMapperBox").value;
    closeCustomMapperModal();
}

// Close modal
function closeCustomMapperModal() {
    document.getElementById("customMapperModal").style.display = "none";
}

// Clear/reset the custom mapper
function clearCustomMapper() {
    customMapper = "";
    document.getElementById("customMapperBox").value = "";
}

// Run button logic
function processInput() {
    const input = inputEditor.getValue();
    const folder = document.getElementById("dropdown1").value;
    const mapper = document.getElementById("dropdown2").value;

    fetch("/process", {
        method: "POST",
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            folder: folder,
            mapper: mapper,
            input: input,
            custom: customMapper
        })
    })
    .then(async res => {
        if (!res.ok) {
            const errorText = await res.text();
            throw new Error(`Server error (${res.status}): ${errorText}`);
        }
        return res.text();
    })
    .then(data => {
        outputEditor.setValue(data);
    })
    .catch(err => {
        console.error("Error:", err);
        outputEditor.setValue("Error: " + err.message);
    });
}
