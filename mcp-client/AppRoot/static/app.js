document.addEventListener("DOMContentLoaded", () => {
    const chatMessages = document.getElementById("chat-messages");
    const chatForm = document.getElementById("chat-form");
    const chatInput = document.getElementById("chat-input");

    function appendTypingIndicator () {
        const typingDiv = document.createElement("div");
        typingDiv.className = "message.bot-response";

        const dots = document.createElement("div");
        dots.className = "typing-dots";
        dots.innerHTML = '<span></span><span></span><span></span>';
 
        typingDiv.appendChild(dots);
        chatMessages.appendChild(typingDiv);
        chatMessages.scrollTop = chatMessages.scrollHeight;

        return typingDiv
    }
 
    function addMessage(role, content) {
        const msgDiv = document.createElement("div");
        msgDiv.classList.add("message", role);
        //msgDiv.textContent = content;  // does not hanlde new line character \n
        // convert common special characters to HTML-friendly formats

        try {
            // attempt to parse the content as JSON
            const expandedContent = content.replace(/\+/g, '"expandable": true,');
            //console.log("JSON content: ", expandedContent)
            const jsonData = JSON.parse(expandedContent);
            msgDiv.innerHTML = `<pre>${JSON.stringify(jsonData, null, 2)}</pre>`
        } catch (error) {
            //console.log("Text content: ", content)
            //const formattedContent = content
            //    .replace(/\n/g, "<br>")                        // new line to <br>
            //    .replace(/\t/g, "&nbsp;&nbsp;&nbsp;&nbsp;")    // tab to Spaces
            //    .replace(/"/g, "&quot;")                       // preserve double quotes
            //    .replace(/'/g, "&#39;")                        // preserve single quotes
            //    .replace(/&/g, "&amp;")                        // preserve &
            //    .replace(/\//g, "&#x2F;")                      // preserve \
            //    .replace(/\\/g, "\\\\")                        // preserve \\

            //msgDiv.innerHTML = document.createElement("textarea").formattedContent
            const textarea = document.createElement("textarea");
            textarea.value = content;
            msgDiv.innerText = textarea.value;
        }
 
        chatMessages.appendChild(msgDiv);
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }
 
    chatForm.addEventListener("submit", async (e) => {
        e.preventDefault();
        const text = chatInput.value.trim();

        if (!text) return;
        // Show user message immediately
        addMessage("user", text);

        let typingIndicator = null;

        try {
            // append typing indicator before sending request
            typingIndicator = appendTypingIndicator();
            chatInput.value = '';
 
            // send request
            const response = await fetch("/chat", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ role: "user", content: text }),
            });
 
            // remove typing indicator after response
            if (typingIndicator) {
                chatMessages.removeChild(typingIndicator);
            }
 
            if (!response.ok) throw new Error("Network response was not ok");

            const data = await response.json();
            if (data.role && data.content) {
                addMessage(data.role, data.content);
            }
        } catch (error) {
            addMessage("assistant", "Error: Unable to get response.");
            console.error("Chat error:", error);
        }
 
        chatInput.value = "";
    });
 
    // New Chat button clears messages and textarea
    document.querySelector('.new-chat-btn').addEventListener('click', () => {
        chatMessages.innerHTML = '';
        chatInput.value = '';
        chatInput.focus();
    });
});
