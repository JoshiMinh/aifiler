const input = document.getElementById('terminal-input');
const output = document.getElementById('terminal-output');
const body = document.getElementById('terminal-body');
const asciiContainer = document.getElementById('ascii-art');

// Initialize ASCII art
asciiContainer.textContent = window.ASCII_ART;

const commands = {
    'help': () => {
        writeLines([
            'Available commands:',
            '  <span class="color-cyan">aifiler "request"</span> - Process a filesystem request',
            '  <span class="color-cyan">aifiler plan</span>      - View the current staged plan',
            '  <span class="color-cyan">aifiler status</span>    - Check AI provider status',
            '  <span class="color-cyan">aifiler undo</span>      - Revert the last change',
            '  <span class="color-cyan">clear</span>             - Clear the screen',
            '  <span class="color-cyan">help</span>              - Show this help menu'
        ]);
    },
    'clear': () => {
        output.innerHTML = '';
        asciiContainer.style.display = 'none';
    },
    'aifiler status': () => {
        writeLines([
            'Checking AI providers...',
            '  [<span class="color-green">OK</span>] OpenAI (gpt-4o)',
            '  [<span class="color-green">OK</span>] Anthropic (claude-3-5-sonnet)',
            '  [<span class="color-yellow">WARN</span>] Gemini (Rate limit approaching)',
            '  [<span class="color-gray">OFF</span>] Ollama (Local server not found)'
        ]);
    },
    'aifiler plan': () => {
        writeLines([
            '<span class="color-yellow">Current Staged Plan:</span>',
            '  1. <span class="color-purple">RENAME</span> "IMG_2024.jpg" -> "Family_Vacation_2024.jpg"',
            '  2. <span class="color-purple">MOVE</span>   "Family_Vacation_2024.jpg" to "/Pictures/2024/"',
            '  3. <span class="color-purple">MKDIR</span>  "/Pictures/2024/"',
            '',
            'Run <span class="color-cyan">aifiler approve</span> to execute.'
        ]);
    }
};

function writeLine(text, delay = 0) {
    setTimeout(() => {
        const line = document.createElement('div');
        line.className = 'output-line';
        line.innerHTML = text;
        output.appendChild(line);
        body.scrollTop = body.scrollHeight;
    }, delay);
}

function writeLines(lines) {
    lines.forEach((line, index) => {
        writeLine(line, index * 80);
    });
}

function handleInput(e) {
    if (e.key === 'Enter') {
        const cmd = input.value.trim().toLowerCase();
        
        // Echo command
        const echo = document.createElement('div');
        echo.className = 'output-line';
        echo.innerHTML = `<span class="prompt-prefix">PS D:\\Projects\\aifiler></span> ${input.value}`;
        output.appendChild(echo);
        
        input.value = '';
        
        if (commands[cmd]) {
            commands[cmd]();
        } else if (cmd.startsWith('aifiler "') || cmd.startsWith('aifiler \'')) {
            simulateAIFiler(cmd);
        } else if (cmd === '') {
            // Do nothing
        } else {
            writeLine(`<span class="color-error">The term '${cmd}' is not recognized as a name of a command.</span>`);
        }
    }
}

function simulateAIFiler(cmd) {
    writeLines([
        '<span class="color-cyan">⚡ Scanning root directory...</span>',
        '<span class="color-cyan">🧠 Consulting AI agent...</span>',
        '   Plan generated successfully.',
        '',
        'Proposed action: <span class="color-yellow">Organize 12 files into categorized subfolders.</span>',
        'Press [Y/n] to approve, or run <span class="color-cyan">aifiler plan</span> for details.'
    ]);
}

window.runDemo = (cmd) => {
    input.value = cmd;
    input.focus();
    const event = new KeyboardEvent('keydown', { key: 'Enter' });
    handleInput(event);
};

// Initial state
input.addEventListener('keydown', handleInput);
document.addEventListener('click', () => input.focus());

// Splash screen
writeLines([
    'Loading <span class="color-cyan">aifiler</span> environment...',
    'Found <span class="color-green">config.yaml</span>',
    'Ready for input.'
]);
