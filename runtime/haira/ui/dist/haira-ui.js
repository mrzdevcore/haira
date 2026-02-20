var R=`
  :host {
    
  --haira-bg: #09090b;
  --haira-bg-card: #0f0f12;
  --haira-bg-card-hover: #18181b;
  --haira-bg-elevated: #1c1c20;
  --haira-bg-input: #0c0c0f;
  --haira-border: rgba(63, 63, 70, 0.5);
  --haira-border-light: rgba(232, 163, 23, 0.12);
  --haira-border-focus: rgba(232, 163, 23, 0.4);
  --haira-accent: #e8a317;
  --haira-accent-light: #f0bd4f;
  --haira-accent-dim: rgba(232, 163, 23, 0.06);
  --haira-glow: #fde68a;
  --haira-text: #fafaf9;
  --haira-text-dim: #a1a1aa;
  --haira-muted: #71717a;
  --haira-success: #22c55e;
  --haira-error: #ef4444;
  --haira-warn: #eab308;
  --haira-info: #3b82f6;
  --haira-radius: 10px;
  --haira-radius-sm: 6px;
  --haira-font: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  --haira-mono: 'SF Mono', 'Fira Code', 'JetBrains Mono', 'Cascadia Code', monospace;

    font-family: var(--haira-font);
    color: var(--haira-text);
    -webkit-font-smoothing: antialiased;
  }
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
`,A=`
  ::-webkit-scrollbar { width: 5px; height: 5px; }
  ::-webkit-scrollbar-track { background: transparent; }
  ::-webkit-scrollbar-thumb { background: var(--haira-muted); border-radius: 3px; }
  ::-webkit-scrollbar-thumb:hover { background: var(--haira-accent); }
  scrollbar-width: thin;
  scrollbar-color: var(--haira-muted) transparent;
`,z=`
  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }
  @keyframes fadeSlideUp {
    from { opacity: 0; transform: translateY(6px); }
    to { opacity: 1; transform: translateY(0); }
  }
  @keyframes pop {
    0% { transform: scale(1); }
    50% { transform: scale(1.02); }
    100% { transform: scale(1); }
  }
  @keyframes spin {
    to { transform: rotate(360deg); }
  }
  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.7; }
  }
  @keyframes blink {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.4; }
  }
  @keyframes bounce {
    0%, 80%, 100% { transform: translateY(0); }
    40% { transform: translateY(-6px); }
  }
  @keyframes expandDown {
    from { opacity: 0; max-height: 0; }
    to { opacity: 1; max-height: 600px; }
  }
`;var V0='<svg width="16" height="16" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-dasharray="28 10" style="animation:spin 0.7s linear infinite;transform-origin:center"/></svg>',G0='<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><path d="M4.5 8.5L7 11L11.5 5.5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>',W0='<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><path d="M5 5L11 11M11 5L5 11" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>',X0='<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><path d="M3 8a5 5 0 0 1 8.5-3.5M13 8a5 5 0 0 1-8.5 3.5" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/><path d="M11 2v3h-3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/><path d="M5 14v-3h3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>',I='<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><rect x="5" y="5" width="9" height="9" rx="1.5" stroke="currentColor" stroke-width="1.5"/><path d="M11 5V3.5A1.5 1.5 0 0 0 9.5 2H3.5A1.5 1.5 0 0 0 2 3.5v6A1.5 1.5 0 0 0 3.5 11H5" stroke="currentColor" stroke-width="1.5"/></svg>',w='<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><path d="M4 8.5L6.5 11L12 5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>',K0='<svg width="12" height="12" viewBox="0 0 16 16" fill="none"><path d="M6 4l4 4-4 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>',O='<svg width="22" height="22" viewBox="0 0 64 52" fill="none" xmlns="http://www.w3.org/2000/svg"><rect x="17" y="11" width="30" height="20" rx="6" fill="#F0BD4F"/><rect x="21" y="15" width="22" height="9" rx="4" fill="#3D2B1F"/><circle cx="27" cy="19.5" r="3.5" fill="#FDE68A"/><circle cx="27" cy="19.5" r="1.5" fill="#fff"/><circle cx="37" cy="19.5" r="3.5" fill="#FDE68A"/><circle cx="37" cy="19.5" r="1.5" fill="#fff"/><ellipse cx="32" cy="12" rx="25" ry="4" fill="#C4A265"/><ellipse cx="32" cy="11.5" rx="23" ry="3" fill="#D4B87A"/><rect x="18" y="2" width="28" height="10" rx="5" fill="#C4A265"/><rect x="20" y="1" width="24" height="5" rx="3" fill="#D4B87A"/><rect x="18" y="8" width="28" height="3.5" rx="1.5" fill="#5C3A1E"/><rect x="20" y="31" width="24" height="14" rx="4" fill="#E8A317"/><rect x="25" y="34" width="14" height="8" rx="3" fill="#3D2B1F"/><circle cx="32" cy="38" r="3" fill="#FDE68A"/><rect x="10" y="34" width="10" height="4" rx="2" fill="#E8A317"/><rect x="44" y="34" width="10" height="4" rx="2" fill="#E8A317"/></svg>';function D(v){switch(v){case"POST":return"#22c55e";case"GET":return"#3b82f6";case"PUT":return"#f59e0b";case"DELETE":return"#ef4444";default:return"#71717a"}}function n(v){switch(v){case"form":return"#3b82f6";case"chat":return"#22c55e";case"stream":return"#e8a317";default:return"#71717a"}}var b0=`
  --haira-bg: #ffffff;
  --haira-bg-card: #f7f7f8;
  --haira-bg-card-hover: #eeeff1;
  --haira-bg-elevated: #e8e8ec;
  --haira-bg-input: #f2f2f4;
  --haira-border: rgba(0, 0, 0, 0.1);
  --haira-border-light: rgba(0, 0, 0, 0.06);
  --haira-border-focus: rgba(0, 0, 0, 0.25);
  --haira-text: #1a1a1a;
  --haira-text-dim: #4a4a4a;
  --haira-muted: #8a8a8a;
`;function A0(v){let i=v.match(/^#?([0-9a-f]{6})$/i);if(!i)return null;return{r:parseInt(i[1].substring(0,2),16),g:parseInt(i[1].substring(2,4),16),b:parseInt(i[1].substring(4,6),16)}}function t0(v,i){let d=A0(v);if(!d)return v;let x=Math.min(255,d.r+Math.round((255-d.r)*i)),$=Math.min(255,d.g+Math.round((255-d.g)*i)),g=Math.min(255,d.b+Math.round((255-d.b)*i));return`#${x.toString(16).padStart(2,"0")}${$.toString(16).padStart(2,"0")}${g.toString(16).padStart(2,"0")}`}class p extends HTMLElement{meta=null;connectedCallback(){let v=document.getElementById("haira-meta");if(v)try{this.meta=JSON.parse(v.textContent||"{}")}catch{this.meta=null}if(this.meta){let i=this.meta.title||this.meta.name;document.title=i?`${i} — Haira`:"Haira"}this.render()}applyTheme(v){if(!this.meta)return;if(this.meta.theme==="light")for(let d of b0.split(`
`)){let x=d.match(/(--[\w-]+):\s*(.+);/);if(x)v.style.setProperty(x[1],x[2].trim())}let i=this.meta.accent;if(i){v.style.setProperty("--haira-accent",i),v.style.setProperty("--haira-accent-light",t0(i,0.25));let d=A0(i);if(d)v.style.setProperty("--haira-accent-dim",`rgba(${d.r}, ${d.g}, ${d.b}, 0.06)`),v.style.setProperty("--haira-border-light",`rgba(${d.r}, ${d.g}, ${d.b}, 0.12)`),v.style.setProperty("--haira-border-focus",`rgba(${d.r}, ${d.g}, ${d.b}, 0.4)`)}}render(){let v=this.attachShadow({mode:"open"});v.innerHTML=`
      <style>
        ${R}
        :host {
          display: block;
          height: 100vh;
          overflow: hidden;
          background: var(--haira-bg);
        }
        .shell {
          height: 100%;
          display: flex;
          flex-direction: column;
          overflow: hidden;
        }
        header {
          padding: 0.6rem 1.25rem;
          border-bottom: 1px solid var(--haira-border);
          display: flex;
          align-items: center;
          gap: 0.6rem;
          background: var(--haira-bg);
          position: sticky;
          top: 0;
          z-index: 100;
        }
        .logo {
          display: flex;
          align-items: center;
          gap: 0.4rem;
          text-decoration: none;
        }
        .logo-icon {
          display: flex;
          align-items: center;
        }
        .logo-icon img {
          width: 22px;
          height: 22px;
          object-fit: contain;
        }
        .logo-text {
          font-weight: 700;
          font-size: 0.92rem;
          color: var(--haira-text);
          letter-spacing: -0.01em;
        }
        .logo-text .ai {
          color: var(--haira-accent);
        }
        .sep {
          color: var(--haira-muted);
          font-size: 0.75rem;
          opacity: 0.5;
        }
        .title {
          color: var(--haira-text-dim);
          font-size: 0.85rem;
          font-weight: 500;
        }
        main {
          flex: 1;
          display: flex;
          flex-direction: column;
          overflow: hidden;
          min-height: 0;
        }
        main.scrollable {
          overflow-y: auto;
        }
      </style>
      <div class="shell">
        <header>
          <a class="logo" href="/_ui/">
            <span class="logo-icon">${this.meta?.logo?`<img src="${this.escapeHtml(this.meta.logo)}" alt="logo">`:O}</span>
            <span class="logo-text">home</span>
          </a>
          ${this.meta&&this.meta.mode!=="index"?`
            <span class="sep">/</span>
            <span class="title">${this.escapeHtml(this.meta.title||this.meta.name||"")}</span>
          `:""}
        </header>
        <main id="content" class="${this.meta?.mode!=="chat"?"scrollable":""}"></main>
      </div>
    `,this.applyTheme(v.host);let i=v.getElementById("content");if(!this.meta){i.innerHTML='<p style="padding:2rem;color:var(--haira-muted)">No workflow metadata found.</p>';return}switch(this.meta.mode){case"index":{let d=document.createElement("haira-index");d.setAttribute("data-meta",JSON.stringify(this.meta)),i.appendChild(d);break}case"form":{let d=document.createElement("haira-form");d.setAttribute("data-meta",JSON.stringify(this.meta)),i.appendChild(d);break}case"chat":{let d=document.createElement("haira-chat");d.setAttribute("data-meta",JSON.stringify(this.meta)),i.appendChild(d);break}}}escapeHtml(v){return v.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;")}}class s extends HTMLElement{_fileValue=null;connectedCallback(){let v=this.getAttribute("name")||"",i=this.getAttribute("type")||"string",d=this.attachShadow({mode:"open"});if(d.innerHTML=`
      <style>
        ${R}
        :host { display: block; margin-bottom: 1rem; }

        .field-label {
          display: block;
          font-weight: 500;
          font-size: 0.82rem;
          color: var(--haira-text-dim);
          margin-bottom: 0.4rem;
        }
        .field-sublabel {
          font-weight: 400;
          color: var(--haira-muted);
          font-size: 0.75rem;
          margin-left: 0.25rem;
        }

        /* Text / Number inputs */
        input[type="text"],
        input[type="number"] {
          width: 100%;
          padding: 0.55rem 0.75rem;
          background: var(--haira-bg-input);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius-sm);
          color: var(--haira-text);
          font-size: 0.88rem;
          font-family: var(--haira-font);
          outline: none;
          transition: border-color 0.15s, box-shadow 0.15s;
        }
        input[type="text"]:focus,
        input[type="number"]:focus {
          border-color: var(--haira-accent);
          box-shadow: 0 0 0 3px rgba(232, 163, 23, 0.08);
        }
        input[type="text"]::placeholder,
        input[type="number"]::placeholder {
          color: var(--haira-muted);
          opacity: 0.6;
        }

        /* Toggle switch for booleans */
        .toggle-row {
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: 0.5rem 0;
        }
        .toggle-label {
          font-size: 0.88rem;
          font-weight: 500;
          color: var(--haira-text-dim);
        }
        .toggle {
          position: relative;
          width: 40px;
          height: 22px;
          flex-shrink: 0;
        }
        .toggle input {
          opacity: 0;
          width: 0;
          height: 0;
          position: absolute;
        }
        .toggle-track {
          position: absolute;
          inset: 0;
          background: var(--haira-bg-elevated);
          border: 1px solid var(--haira-border);
          border-radius: 11px;
          cursor: pointer;
          transition: background 0.2s, border-color 0.2s;
        }
        .toggle-track::after {
          content: "";
          position: absolute;
          top: 2px;
          left: 2px;
          width: 16px;
          height: 16px;
          background: var(--haira-muted);
          border-radius: 50%;
          transition: transform 0.2s ease, background 0.2s;
        }
        .toggle input:checked + .toggle-track {
          background: rgba(232, 163, 23, 0.15);
          border-color: var(--haira-accent);
        }
        .toggle input:checked + .toggle-track::after {
          transform: translateX(18px);
          background: var(--haira-accent);
        }
        .toggle input:focus-visible + .toggle-track {
          box-shadow: 0 0 0 3px rgba(232, 163, 23, 0.15);
        }

        /* File drop zone */
        .drop-zone {
          position: relative;
          border: 1.5px dashed var(--haira-border);
          border-radius: var(--haira-radius);
          padding: 1.25rem 1rem;
          text-align: center;
          cursor: pointer;
          transition: all 0.2s;
          background: var(--haira-bg-input);
        }
        .drop-zone:hover,
        .drop-zone.dragover {
          border-color: var(--haira-accent);
          background: rgba(232, 163, 23, 0.03);
        }
        .drop-zone.has-file {
          border-style: solid;
          border-color: var(--haira-success);
          background: rgba(34, 197, 94, 0.03);
        }
        .drop-icon {
          margin-bottom: 0.35rem;
          color: var(--haira-muted);
        }
        .drop-zone.has-file .drop-icon { color: var(--haira-success); }
        .drop-text {
          font-size: 0.82rem;
          color: var(--haira-muted);
          line-height: 1.4;
        }
        .drop-text strong {
          color: var(--haira-accent);
          cursor: pointer;
        }
        .drop-zone.has-file .drop-text { color: var(--haira-text-dim); }
        .drop-text .filename {
          display: block;
          font-family: var(--haira-mono);
          font-size: 0.8rem;
          color: var(--haira-text);
          margin-top: 0.2rem;
          word-break: break-all;
        }
        .drop-text .filesize {
          font-size: 0.72rem;
          color: var(--haira-muted);
        }
        .drop-zone input[type="file"] {
          position: absolute;
          inset: 0;
          width: 100%;
          height: 100%;
          opacity: 0;
          cursor: pointer;
        }
        .clear-btn {
          display: none;
          position: absolute;
          top: 0.5rem;
          right: 0.5rem;
          background: var(--haira-bg-elevated);
          border: 1px solid var(--haira-border);
          border-radius: 4px;
          color: var(--haira-muted);
          font-size: 0.7rem;
          padding: 0.15rem 0.4rem;
          cursor: pointer;
          transition: all 0.15s;
        }
        .clear-btn:hover { color: var(--haira-error); border-color: var(--haira-error); }
        .drop-zone.has-file .clear-btn { display: block; }
      </style>
      ${this.renderInput(v,i)}
    `,i==="file")this.setupFileDrop(d)}renderInput(v,i){switch(i){case"bool":return`
          <div class="toggle-row">
            <span class="toggle-label">${this.esc(v)}</span>
            <label class="toggle">
              <input type="checkbox" id="f-${v}" name="${v}">
              <span class="toggle-track"></span>
            </label>
          </div>`;case"file":return`
          <label class="field-label">${this.esc(v)}</label>
          <div class="drop-zone" id="drop-zone">
            <input type="file" id="f-${v}" name="${v}">
            <div class="drop-icon">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none"><path d="M12 16V4m0 0l-4 4m4-4l4 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><path d="M4 17v2a2 2 0 002 2h12a2 2 0 002-2v-2" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
            </div>
            <div class="drop-text" id="drop-text">
              Drop file here or <strong>browse</strong>
            </div>
            <button class="clear-btn" id="clear-btn" type="button">Clear</button>
          </div>`;case"int":return`
          <label class="field-label" for="f-${v}">${this.esc(v)} <span class="field-sublabel">integer</span></label>
          <input type="number" id="f-${v}" name="${v}" step="1" placeholder="0">`;case"float":return`
          <label class="field-label" for="f-${v}">${this.esc(v)} <span class="field-sublabel">number</span></label>
          <input type="number" id="f-${v}" name="${v}" step="any" placeholder="0.0">`;default:return`
          <label class="field-label" for="f-${v}">${this.esc(v)}</label>
          <input type="text" id="f-${v}" name="${v}" placeholder="Enter value...">`}}setupFileDrop(v){let i=v.getElementById("drop-zone"),d=v.querySelector("input[type=file]"),x=v.getElementById("drop-text"),$=v.getElementById("clear-btn");if(!i||!d||!x||!$)return;let g=(k)=>{this._fileValue=k,i.classList.add("has-file");let T=k.size<1024?`${k.size} B`:k.size<1048576?`${(k.size/1024).toFixed(1)} KB`:`${(k.size/1048576).toFixed(1)} MB`;x.innerHTML=`<span class="filename">${this.esc(k.name)}</span><span class="filesize">${T}</span>`},r=()=>{this._fileValue=null,d.value="",i.classList.remove("has-file"),x.innerHTML="Drop file here or <strong>browse</strong>"};d.addEventListener("change",()=>{if(d.files?.[0])g(d.files[0])}),$.addEventListener("click",(k)=>{k.stopPropagation(),r()}),i.addEventListener("dragover",(k)=>{k.preventDefault(),i.classList.add("dragover")}),i.addEventListener("dragleave",()=>{i.classList.remove("dragover")}),i.addEventListener("drop",(k)=>{k.preventDefault(),i.classList.remove("dragover");let T=k.dataTransfer?.files?.[0];if(T){g(T);let c=new DataTransfer;c.items.add(T),d.files=c.files}})}getValue(){let v=this.getAttribute("name")||"",i=this.getAttribute("type")||"string";if(i==="bool"){let x=this.shadowRoot.querySelector("input[type=checkbox]");return{name:v,value:x.checked,type:i}}if(i==="file")return{name:v,value:this._fileValue||this.shadowRoot.querySelector("input[type=file]")?.files?.[0]||null,type:i};let d=this.shadowRoot.querySelector("input");if(i==="int"||i==="float")return{name:v,value:d.value?Number(d.value):"",type:i};return{name:v,value:d.value,type:i}}esc(v){return v.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;")}}class e extends HTMLElement{rawText="";connectedCallback(){let v=this.attachShadow({mode:"open"});v.innerHTML=`
      <style>
        ${R}
        ${z}
        :host { display: none; margin-top: 0.75rem; }
        :host([visible]) {
          display: block;
          animation: fadeSlideUp 0.25s ease-out;
        }
        .card {
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          overflow: hidden;
        }
        .header {
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: 0.6rem 0.85rem;
          border-bottom: 1px solid var(--haira-border);
        }
        .header-left {
          display: flex;
          align-items: center;
          gap: 0.4rem;
          font-weight: 600;
          font-size: 0.78rem;
          color: var(--haira-muted);
        }
        .dot {
          width: 6px;
          height: 6px;
          border-radius: 50%;
          flex-shrink: 0;
        }
        .dot.success { background: var(--haira-success); }
        .dot.error { background: var(--haira-error); }
        .copy-btn {
          background: none;
          border: 1px solid transparent;
          border-radius: 4px;
          padding: 0.25rem;
          cursor: pointer;
          color: var(--haira-muted);
          display: flex;
          align-items: center;
          justify-content: center;
          transition: all 0.15s;
        }
        .copy-btn:hover {
          color: var(--haira-accent);
          border-color: var(--haira-border);
          background: var(--haira-bg-elevated);
        }
        .body {
          padding: 0.85rem;
          font-size: 0.82rem;
          line-height: 1.6;
          color: var(--haira-text-dim);
          max-height: 600px;
          overflow-y: auto;
          ${A}
        }
        /* Rich result styles */
        .body.rich {
          white-space: normal;
          word-break: break-word;
          font-family: var(--haira-font);
        }
        .body.raw {
          white-space: pre-wrap;
          word-break: break-word;
          font-family: var(--haira-mono);
          font-size: 0.8rem;
        }
        .result-section {
          margin-bottom: 0.75rem;
        }
        .result-section:last-child {
          margin-bottom: 0;
        }
        .section-label {
          font-size: 0.68rem;
          font-weight: 700;
          text-transform: uppercase;
          letter-spacing: 0.05em;
          color: var(--haira-muted);
          margin-bottom: 0.3rem;
        }
        .section-label.error {
          color: var(--haira-error);
        }
        .section-value {
          color: var(--haira-text);
          line-height: 1.55;
        }
        .section-value ul {
          margin: 0.25rem 0 0 0;
          padding-left: 1.25rem;
        }
        .section-value li {
          margin-bottom: 0.15rem;
          font-family: var(--haira-mono);
          font-size: 0.78rem;
          color: var(--haira-text-dim);
        }
        .code-block {
          background: var(--haira-bg);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius-sm);
          padding: 0.65rem 0.85rem;
          margin-top: 0.35rem;
          font-family: var(--haira-mono);
          font-size: 0.75rem;
          line-height: 1.5;
          white-space: pre-wrap;
          word-break: break-all;
          overflow-x: auto;
          color: var(--haira-text);
        }
        .code-lang {
          font-size: 0.62rem;
          text-transform: uppercase;
          color: var(--haira-muted);
          letter-spacing: 0.04em;
          margin-bottom: 0.2rem;
          font-weight: 600;
        }
        .result-kv {
          display: flex;
          gap: 0.5rem;
          padding: 0.2rem 0;
        }
        .result-kv .kv-key {
          font-size: 0.72rem;
          font-weight: 600;
          color: var(--haira-muted);
          min-width: 60px;
          flex-shrink: 0;
        }
        .result-kv .kv-val {
          color: var(--haira-text);
          font-size: 0.82rem;
        }
      </style>
      <div class="card">
        <div class="header">
          <div class="header-left">
            <span class="dot" id="dot"></span>
            <span id="label">Result</span>
          </div>
          <button class="copy-btn" id="copy-btn" title="Copy to clipboard">${I}</button>
        </div>
        <div class="body raw" id="body"></div>
      </div>
    `,v.getElementById("copy-btn").addEventListener("click",()=>this.copyResult())}show(v,i){this.setAttribute("visible","");let d=this.shadowRoot.getElementById("body"),x=this.shadowRoot.getElementById("dot"),$=this.shadowRoot.getElementById("label");x.className=`dot ${i?"error":"success"}`,$.textContent=i?"Error":"Result";let g=v;if(typeof v==="object"&&v!==null&&typeof g.message==="string"&&g.message.length>0){this.rawText=g.message,d.className="body rich",d.innerHTML=this.renderMessage(g.message,g.status);return}if(typeof v==="object"&&v!==null&&!Array.isArray(v)){let k=Object.keys(g);if(k.length>0&&k.length<=10&&k.every((T)=>typeof g[T]!=="object"||g[T]===null)){this.rawText=JSON.stringify(v,null,2),d.className="body rich",d.innerHTML=k.map((T)=>`<div class="result-kv"><span class="kv-key">${this.esc(T)}</span><span class="kv-val">${this.esc(String(g[T]??""))}</span></div>`).join("");return}}let r;if(typeof v==="string")r=v;else r=JSON.stringify(v,null,2);this.rawText=r,d.className="body raw",d.textContent=r}hide(){this.removeAttribute("visible")}renderMessage(v,i){let d=v.split(`
`),x=[],$=0;while($<d.length){let g=d[$],r=g.match(/^```(\w*)$/);if(r){let T=r[1]||"",c=[];$++;while($<d.length&&!d[$].startsWith("```"))c.push(d[$]),$++;$++,x.push({type:"code",lang:T,content:c.join(`
`)});continue}let k=g.match(/^([A-Z][A-Z _]{2,}):(.*)$/);if(k){let T=k[1].trim(),c=k[2].trim(),L=c?[c]:[];$++;while($<d.length){let W=d[$];if(W.match(/^[A-Z][A-Z _]{2,}:/)||W.startsWith("```"))break;L.push(W),$++}let o=L.join(`
`).trim(),B=o.split(`
`);if(B.length>1&&B.every((W)=>W.startsWith("- ")||W.trim()===""))x.push({type:"list",label:T,content:o});else x.push({type:"heading",label:T,content:o});continue}if(g.trim())x.push({type:"text",content:g});$++}if(x.length===0)return`<div class="section-value">${this.esc(v)}</div>`;return x.map((g)=>{switch(g.type){case"heading":return`<div class="result-section">
              <div class="section-label${i==="error"&&g.label?.includes("CAUSE")?" error":""}">${this.esc(g.label||"")}</div>
              <div class="section-value">${this.esc(g.content)}</div>
            </div>`;case"list":return`<div class="result-section">
              <div class="section-label">${this.esc(g.label||"")}</div>
              <div class="section-value"><ul>${g.content.split(`
`).filter((r)=>r.startsWith("- ")).map((r)=>`<li>${this.esc(r.slice(2))}</li>`).join("")}</ul></div>
            </div>`;case"code":return`<div class="result-section">
              ${g.lang?`<div class="code-lang">${this.esc(g.lang)}</div>`:""}
              <div class="code-block">${this.esc(g.content)}</div>
            </div>`;case"text":return`<div class="section-value">${this.esc(g.content)}</div>`;default:return""}}).join("")}esc(v){return v.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;")}async copyResult(){let v=this.shadowRoot?.getElementById("copy-btn");if(!v)return;try{await navigator.clipboard.writeText(this.rawText),v.innerHTML=w,setTimeout(()=>{v.innerHTML=I},1500)}catch{}}}class v0 extends HTMLElement{_status="pending";_duration;_timerInterval=null;_timerStart=0;_expanded=!1;_logCount=0;_hasError=!1;connectedCallback(){this.render()}disconnectedCallback(){this.clearTimer()}render(){let v=this.getAttribute("name")||"",i=this.getAttribute("index")||"0",d=this.attachShadow({mode:"open"});d.innerHTML=`
      <style>
        ${R}
        ${z}
        :host {
          display: block;
          position: relative;
        }

        /* --- Step header row (clickable) --- */
        .step-header {
          display: flex;
          align-items: center;
          gap: 0.6rem;
          padding: 0.5rem 0.65rem;
          border-radius: var(--haira-radius-sm);
          cursor: pointer;
          user-select: none;
          transition: background 0.15s;
          position: relative;
        }
        .step-header:hover {
          background: rgba(255, 255, 255, 0.03);
        }

        /* Chevron */
        .chevron {
          flex-shrink: 0;
          width: 16px;
          height: 16px;
          display: flex;
          align-items: center;
          justify-content: center;
          color: var(--haira-muted);
          transition: transform 0.2s ease, color 0.2s;
          opacity: 0;
        }
        .has-logs .chevron { opacity: 1; }
        .expanded .chevron {
          transform: rotate(90deg);
        }

        /* Status indicator */
        .status-icon {
          flex-shrink: 0;
          width: 22px;
          height: 22px;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          transition: all 0.25s ease;
        }
        .pending .status-icon {
          border: 1.5px dashed var(--haira-muted);
          color: var(--haira-muted);
        }
        .running .status-icon {
          border: 1.5px solid var(--haira-accent);
          color: var(--haira-accent);
          background: rgba(232, 163, 23, 0.1);
        }
        .done .status-icon {
          background: var(--haira-success);
          color: #fff;
        }
        .failed .status-icon {
          background: var(--haira-error);
          color: #fff;
        }
        .retrying .status-icon {
          border: 1.5px solid var(--haira-accent);
          color: var(--haira-accent);
          background: rgba(232, 163, 23, 0.1);
          animation: pulse 1.5s ease-in-out infinite;
        }
        .skipped .status-icon {
          border: 1.5px dashed var(--haira-muted);
          color: var(--haira-muted);
          opacity: 0.5;
        }

        .step-num {
          font-size: 0.65rem;
          font-weight: 600;
        }

        /* Step name */
        .step-name {
          flex: 1;
          font-size: 0.85rem;
          font-weight: 500;
          color: var(--haira-muted);
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
          transition: color 0.2s;
        }
        .running .step-name { color: var(--haira-text); font-weight: 600; }
        .done .step-name { color: var(--haira-text-dim); }
        .failed .step-name { color: var(--haira-text); }
        .retrying .step-name { color: var(--haira-text); }

        /* Log count badge */
        .log-count {
          font-size: 0.7rem;
          color: var(--haira-muted);
          padding: 0.1rem 0.4rem;
          border-radius: 10px;
          background: rgba(255, 255, 255, 0.04);
          font-family: var(--haira-mono);
          display: none;
        }
        .has-logs .log-count { display: inline-block; }
        .has-error .log-count {
          color: var(--haira-error);
          background: rgba(239, 68, 68, 0.1);
        }

        /* Timer */
        .timer {
          flex-shrink: 0;
          font-size: 0.75rem;
          font-family: var(--haira-mono);
          color: var(--haira-muted);
          min-width: 36px;
          text-align: right;
          transition: color 0.2s;
        }
        .running .timer { color: var(--haira-accent-light); }
        .done .timer { color: var(--haira-success); }
        .failed .timer { color: var(--haira-error); }

        /* --- Collapsible log area --- */
        .logs-wrapper {
          overflow: hidden;
          max-height: 0;
          opacity: 0;
          transition: max-height 0.25s ease, opacity 0.2s ease;
          margin-left: 2.55rem;
        }
        .logs-wrapper.open {
          max-height: 600px;
          opacity: 1;
          overflow-y: auto;
          ${A}
        }
        .logs-inner {
          padding: 0.25rem 0 0.5rem 0;
          border-left: 1px solid rgba(63, 63, 70, 0.3);
          margin-left: 0.15rem;
        }
        .log-entry {
          display: flex;
          align-items: flex-start;
          gap: 0.5rem;
          font-size: 0.78rem;
          font-family: var(--haira-mono);
          line-height: 1.5;
          padding: 0.15rem 0 0.15rem 0.85rem;
          animation: fadeIn 0.15s ease-out both;
        }
        .log-badge {
          flex-shrink: 0;
          font-size: 0.6rem;
          font-weight: 700;
          text-transform: uppercase;
          padding: 0.08rem 0.35rem;
          border-radius: 3px;
          letter-spacing: 0.04em;
          margin-top: 0.12rem;
        }
        .log-badge.info {
          background: rgba(59, 130, 246, 0.12);
          color: var(--haira-info);
        }
        .log-badge.warn {
          background: rgba(234, 179, 8, 0.12);
          color: var(--haira-warn);
        }
        .log-badge.error {
          background: rgba(239, 68, 68, 0.12);
          color: var(--haira-error);
        }
        .log-msg {
          flex: 1;
          word-break: break-word;
          white-space: pre-wrap;
          color: var(--haira-text-dim);
        }
        .log-msg.warn { color: var(--haira-warn); }
        .log-msg.error { color: var(--haira-error); }

        /* Error detail */
        .error-detail {
          margin: 0.25rem 0 0.5rem 0;
          padding: 0.5rem 0.75rem;
          font-size: 0.78rem;
          font-family: var(--haira-mono);
          color: var(--haira-error);
          background: rgba(239, 68, 68, 0.06);
          border: 1px solid rgba(239, 68, 68, 0.12);
          border-radius: var(--haira-radius-sm);
          margin-left: 2.55rem;
          line-height: 1.5;
          word-break: break-word;
          white-space: pre-wrap;
          display: none;
        }
        .error-detail.visible { display: block; animation: fadeIn 0.2s ease-out; }

        @keyframes spin { to { transform: rotate(360deg); } }
      </style>
      <div class="step-header pending" id="header">
        <span class="chevron" id="chevron">${K0}</span>
        <span class="status-icon" id="status-icon">
          <span class="step-num" id="step-num">${Number(i)+1}</span>
        </span>
        <span class="step-name">${this.esc(v)}</span>
        <span class="log-count" id="log-count"></span>
        <span class="timer" id="timer"></span>
      </div>
      <div class="logs-wrapper" id="logs-wrapper">
        <div class="logs-inner" id="logs"></div>
      </div>
      <div class="error-detail" id="error-detail"></div>
    `,d.getElementById("header").addEventListener("click",()=>{if(this._logCount===0)return;this.toggleLogs()})}toggleLogs(v){let i=this.shadowRoot?.getElementById("logs-wrapper"),d=this.shadowRoot?.getElementById("header");if(!i||!d)return;if(v===!0)this._expanded=!0;else if(v===!1)this._expanded=!1;else this._expanded=!this._expanded;if(this._expanded)i.classList.add("open"),d.classList.add("expanded");else i.classList.remove("open"),d.classList.remove("expanded")}setStatus(v,i,d){this._status=v,this._duration=i;let x=this.shadowRoot?.getElementById("header"),$=this.shadowRoot?.getElementById("status-icon"),g=this.shadowRoot?.getElementById("timer"),r=this.shadowRoot?.getElementById("error-detail");if(!x||!$||!g)return;let k=[];if(this._logCount>0)k.push("has-logs");if(this._hasError)k.push("has-error");if(this._expanded)k.push("expanded");switch(x.className=`step-header ${v} ${k.join(" ")}`,v){case"pending":$.innerHTML=`<span class="step-num">${this.getIndex()}</span>`,this.clearTimer(),g.textContent="",r?.classList.remove("visible");break;case"running":$.innerHTML=V0,this.startTimer(g),r?.classList.remove("visible");break;case"done":if($.innerHTML=G0,this.clearTimer(),g.textContent=this.formatDuration(i),r?.classList.remove("visible"),!this._hasError)this.toggleLogs(!1);break;case"failed":if($.innerHTML=W0,this.clearTimer(),g.textContent=this.formatDuration(i),d&&r)r.textContent=d,r.classList.add("visible");if(this._logCount>0)this.toggleLogs(!0);break;case"retrying":$.innerHTML=X0,r?.classList.remove("visible");break;case"skipped":$.innerHTML=`<span class="step-num">${this.getIndex()}</span>`,this.clearTimer(),g.textContent="skipped",r?.classList.remove("visible");break}}addLog(v,i){let d=this.shadowRoot?.getElementById("logs"),x=this.shadowRoot?.getElementById("header"),$=this.shadowRoot?.getElementById("log-count");if(!d||!x||!$)return;if(this._logCount++,x.classList.add("has-logs"),$.textContent=String(this._logCount),v==="error")this._hasError=!0,x.classList.add("has-error"),this.toggleLogs(!0);let g=200,r=i.length>g?i.slice(0,g)+"...":i,k=document.createElement("div");if(k.className="log-entry",k.innerHTML=`<span class="log-badge ${v}">${v}</span><span class="log-msg ${v}">${this.esc(r)}</span>`,d.appendChild(k),this._status==="running"&&!this._expanded)this.toggleLogs(!0);let T=this.shadowRoot?.getElementById("logs-wrapper");if(T&&this._expanded)T.scrollTop=T.scrollHeight}clearLogs(){let v=this.shadowRoot?.getElementById("logs"),i=this.shadowRoot?.getElementById("header"),d=this.shadowRoot?.getElementById("error-detail");if(!v)return;v.innerHTML="",this._logCount=0,this._hasError=!1,this._expanded=!1,i?.classList.remove("has-logs","has-error","expanded"),d?.classList.remove("visible"),this.toggleLogs(!1);let x=this.shadowRoot?.getElementById("log-count");if(x)x.textContent=""}getIndex(){let v=this.getAttribute("index");return v!==null?String(Number(v)+1):""}startTimer(v){this.clearTimer(),this._timerStart=performance.now(),v.textContent="0.0s",this._timerInterval=setInterval(()=>{let i=(performance.now()-this._timerStart)/1000;v.textContent=`${i.toFixed(1)}s`},100)}clearTimer(){if(this._timerInterval!==null)clearInterval(this._timerInterval),this._timerInterval=null}formatDuration(v){if(v===void 0)return"";return v<1000?`${v}ms`:`${(v/1000).toFixed(1)}s`}get status(){return this._status}get duration(){return this._duration}esc(v){return v.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;")}}class i0 extends HTMLElement{steps=[];stepElements=[];stepStatuses=[];totalDuration=0;connectedCallback(){let v=this.attachShadow({mode:"open"});v.innerHTML=`
      <style>
        ${R}
        ${z}
        :host {
          display: none;
          margin-top: 1.25rem;
        }
        :host([visible]) {
          display: block;
          animation: fadeIn 0.2s ease-out;
        }
        .header {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          padding: 0 0.25rem 0.6rem;
          font-size: 0.72rem;
          font-weight: 600;
          color: var(--haira-muted);
          text-transform: uppercase;
          letter-spacing: 0.06em;
        }
        .header-line {
          flex: 1;
          height: 1px;
          background: var(--haira-border);
        }
        .pipeline {
          display: flex;
          flex-direction: column;
          gap: 1px;
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          padding: 0.35rem;
          overflow: hidden;
        }
        .summary {
          display: none;
          padding: 0.75rem 1rem;
          font-size: 0.78rem;
          color: var(--haira-muted);
          border-top: 1px solid var(--haira-border);
          margin-top: 0.5rem;
          border-radius: var(--haira-radius-sm);
          background: var(--haira-bg-card);
          text-align: center;
        }
        .summary.visible {
          display: block;
          animation: fadeIn 0.3s ease-out;
        }
        .summary .count {
          color: var(--haira-text-dim);
          font-weight: 500;
        }
        .summary .time {
          color: var(--haira-accent);
          font-family: var(--haira-mono);
          font-weight: 600;
        }
        .summary .failed-count {
          color: var(--haira-error);
        }
      </style>
      <div class="header">
        <span>Pipeline</span>
        <span class="header-line"></span>
      </div>
      <div class="pipeline" id="pipeline"></div>
      <div class="summary" id="summary"></div>
    `}setSteps(v){this.steps=v,this.stepElements=[],this.stepStatuses=v.map(()=>"pending"),this.totalDuration=0;let i=this.shadowRoot?.getElementById("pipeline"),d=this.shadowRoot?.getElementById("summary");if(!i)return;if(i.innerHTML="",d)d.classList.remove("visible"),d.textContent="";v.forEach((x,$)=>{let g=document.createElement("haira-step");g.setAttribute("name",x),g.setAttribute("index",String($)),i.appendChild(g),this.stepElements.push(g)})}updateStep(v){let i=this.steps.indexOf(v.name);if(i===-1)return;let d=this.stepElements[i];if(!d)return;if(v.status==="log"&&v.log){d.addLog(v.log.level,v.log.message);return}let x;switch(v.status){case"start":x="running";break;case"end":if(x="done",v.duration_ms)this.totalDuration+=v.duration_ms;break;case"failed":x="failed";break;case"retry":x="retrying";break;default:return}this.stepStatuses[i]=x,d.setStatus(x,v.duration_ms,v.error),this.checkCompletion()}checkCompletion(){if(!this.stepStatuses.every((k)=>k==="done"||k==="failed"||k==="skipped"))return;let i=this.shadowRoot?.getElementById("summary");if(!i)return;let d=this.stepStatuses.filter((k)=>k==="done").length,x=this.stepStatuses.filter((k)=>k==="failed").length,$=this.stepStatuses.filter((k)=>k==="skipped").length,g=(this.totalDuration/1000).toFixed(1),r=`<span class="count">${d}/${this.steps.length} steps completed`;if(x>0)r+=` &middot; <span class="failed-count">${x} failed</span>`;if($>0)r+=` &middot; ${$} skipped`;r+=`</span> &middot; <span class="time">${g}s total</span>`,i.innerHTML=r,i.classList.add("visible")}finalize(){for(let v=0;v<this.stepStatuses.length;v++){let i=this.stepStatuses[v];if(i==="running")this.stepStatuses[v]="done",this.stepElements[v].setStatus("done");else if(i==="pending")this.stepStatuses[v]="skipped",this.stepElements[v].setStatus("skipped")}this.checkCompletion()}reset(){this.totalDuration=0,this.stepStatuses=this.steps.map(()=>"pending");for(let i of this.stepElements)i.clearLogs(),i.setStatus("pending");let v=this.shadowRoot?.getElementById("summary");if(v)v.classList.remove("visible"),v.innerHTML=""}show(){this.setAttribute("visible","")}hide(){this.removeAttribute("visible")}}var a0='<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>';class d0 extends HTMLElement{connectedCallback(){this.render()}render(){let v=this.getAttribute("role")||"user",i=this.getAttribute("content")||"",d=this.getAttribute("file")||"",x=this.getAttribute("avatar")||"H",$=this.attachShadow({mode:"open"});$.innerHTML=`
      <style>
        ${R}
        ${z}
        :host {
          display: block;
          animation: fadeSlideUp 0.2s ease-out;
        }
        .row {
          display: flex;
          gap: 0.6rem;
          align-items: flex-start;
        }
        .row.user {
          justify-content: flex-end;
        }
        .row.assistant {
          justify-content: flex-start;
        }

        /* Avatar */
        .avatar {
          width: 26px;
          height: 26px;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          flex-shrink: 0;
          font-size: 0.65rem;
          font-weight: 700;
          margin-top: 2px;
        }
        .avatar.assistant {
          background: var(--haira-accent-dim);
          border: 1px solid rgba(232, 163, 23, 0.2);
          color: var(--haira-accent);
        }
        .avatar.assistant img {
          width: 100%;
          height: 100%;
          border-radius: 50%;
          object-fit: cover;
        }
        .avatar.user {
          display: none;
        }

        /* Bubble */
        .bubble {
          padding: 0.7rem 0.9rem;
          border-radius: 12px;
          line-height: 1.6;
          font-size: 0.88rem;
          word-wrap: break-word;
          min-width: 40px;
        }
        .bubble.user {
          background: var(--haira-bg-elevated);
          border: 1px solid var(--haira-border);
          color: var(--haira-text);
          border-bottom-right-radius: 4px;
          max-width: 85%;
        }
        .bubble.assistant {
          background: transparent;
          color: var(--haira-text);
          padding: 0.4rem 0;
          flex: 1;
        }

        /* File chip in user message */
        .file-tag {
          display: inline-flex;
          align-items: center;
          gap: 0.3rem;
          background: rgba(0,0,0,0.12);
          padding: 0.2rem 0.5rem;
          border-radius: 5px;
          font-size: 0.75rem;
          margin-bottom: 0.3rem;
          font-weight: 500;
        }
        .file-tag svg { opacity: 0.7; }

        /* Assistant markdown */
        .bubble.assistant .code-wrapper {
          position: relative;
          margin: 0.5rem 0;
        }
        .bubble.assistant pre {
          background: var(--haira-bg);
          border: 1px solid var(--haira-border);
          padding: 0.6rem 0.75rem;
          border-radius: 6px;
          overflow-x: auto;
          font-size: 0.78rem;
          font-family: var(--haira-mono);
          line-height: 1.5;
          margin: 0;
        }
        .bubble.assistant code {
          background: var(--haira-bg-elevated);
          border: 1px solid var(--haira-border);
          padding: 0.1rem 0.3rem;
          border-radius: 3px;
          font-size: 0.8rem;
          font-family: var(--haira-mono);
          color: var(--haira-accent-light);
        }
        .bubble.assistant pre code {
          background: none;
          border: none;
          padding: 0;
          color: var(--haira-text);
        }
        .bubble.assistant strong { font-weight: 700; }
        .bubble.assistant em { font-style: italic; color: var(--haira-text-dim); }
        .bubble.assistant p {
          margin: 0.3rem 0;
        }
        .bubble.assistant p:first-child { margin-top: 0; }
        .bubble.assistant p:last-child { margin-bottom: 0; }
        .bubble.assistant ul, .bubble.assistant ol {
          margin: 0.35rem 0;
          padding-left: 1.3rem;
        }
        .bubble.assistant li {
          margin: 0.2rem 0;
        }
        .bubble.assistant h1, .bubble.assistant h2, .bubble.assistant h3 {
          font-size: 0.9rem;
          font-weight: 700;
          margin: 0.6rem 0 0.25rem;
          color: var(--haira-text);
        }
        .bubble.assistant h1:first-child,
        .bubble.assistant h2:first-child,
        .bubble.assistant h3:first-child {
          margin-top: 0;
        }
        .bubble.assistant hr {
          border: none;
          border-top: 1px solid var(--haira-border);
          margin: 0.5rem 0;
        }
        .bubble.assistant a {
          color: var(--haira-accent);
          text-decoration: none;
        }
        .bubble.assistant a:hover { text-decoration: underline; }
        .bubble.assistant blockquote {
          border-left: 3px solid var(--haira-accent);
          margin: 0.4rem 0;
          padding: 0.2rem 0.6rem;
          color: var(--haira-text-dim);
        }
        .bubble.assistant table {
          border-collapse: collapse;
          width: 100%;
          margin: 0.5rem 0;
          font-size: 0.82rem;
        }
        .bubble.assistant th {
          text-align: left;
          padding: 0.4rem 0.6rem;
          border-bottom: 2px solid var(--haira-border);
          font-weight: 600;
          color: var(--haira-text);
          font-size: 0.78rem;
        }
        .bubble.assistant td {
          padding: 0.35rem 0.6rem;
          border-bottom: 1px solid var(--haira-border);
          color: var(--haira-text-dim);
        }
        .bubble.assistant tr:last-child td {
          border-bottom: none;
        }
        .bubble.assistant ol {
          margin: 0.35rem 0;
          padding-left: 1.3rem;
        }
        .bubble.assistant ol li {
          margin: 0.2rem 0;
        }

        /* Copy button on code blocks */
        .copy-code {
          position: absolute;
          top: 0.4rem;
          right: 0.4rem;
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: 4px;
          padding: 0.2rem;
          cursor: pointer;
          color: var(--haira-muted);
          display: flex;
          align-items: center;
          opacity: 0;
          transition: opacity 0.15s;
        }
        .code-wrapper:hover .copy-code { opacity: 1; }
        .copy-code:hover {
          color: var(--haira-accent);
          border-color: var(--haira-accent);
        }
      </style>
      <div class="row ${v}">
        ${v==="assistant"?`<div class="avatar assistant">${x.startsWith("http")?`<img src="${this.esc(x)}" alt="">`:this.esc(x)}</div>`:""}
        <div class="bubble ${v}" id="bubble"></div>
      </div>
    `;let g=$.getElementById("bubble");if(v==="assistant")g.innerHTML=this.renderMarkdown(i),this.attachCopyHandlers($);else{let r="";if(d)r+=`<div class="file-tag">${a0} ${this.esc(d)}</div><br>`;if(i)r+=this.esc(i);g.innerHTML=r}}updateContent(v){let i=this.shadowRoot?.getElementById("bubble");if(i)i.innerHTML=this.renderMarkdown(v),this.attachCopyHandlers(this.shadowRoot)}attachCopyHandlers(v){v.querySelectorAll(".copy-code").forEach((i)=>{i.addEventListener("click",async()=>{let x=i.closest(".code-wrapper")?.querySelector("code");if(!x)return;try{await navigator.clipboard.writeText(x.textContent||""),i.innerHTML=w,setTimeout(()=>{i.innerHTML=I},1500)}catch{}})})}renderMarkdown(v){let i=this.esc(v);if(i=i.replace(/```(\w*)\n([\s\S]*?)```/g,(d,x,$)=>`<div class="code-wrapper"><pre><code>${$.trim()}</code></pre><button class="copy-code" title="Copy code">${I}</button></div>`),i=i.replace(/((?:^|\n)\|.+\|(?:\n\|[-:| ]+\|)(?:\n\|.+\|)+)/g,(d)=>{let x=d.trim().split(`
`);if(x.length<2)return d;let $=x[0].split("|").filter((k)=>k.trim()),g=x.slice(2),r="<table><thead><tr>";for(let k of $)r+=`<th>${k.trim()}</th>`;r+="</tr></thead><tbody>";for(let k of g){let T=k.split("|").filter((c)=>c.trim());r+="<tr>";for(let c of T)r+=`<td>${c.trim()}</td>`;r+="</tr>"}return r+="</tbody></table>",r}),i=i.replace(/`([^`]+)`/g,"<code>$1</code>"),i=i.replace(/^### (.+)$/gm,"<h3>$1</h3>"),i=i.replace(/^## (.+)$/gm,"<h2>$1</h2>"),i=i.replace(/^# (.+)$/gm,"<h1>$1</h1>"),i=i.replace(/^---$/gm,"<hr>"),i=i.replace(/\*\*(.+?)\*\*/g,"<strong>$1</strong>"),i=i.replace(/(?<!\w)\*(.+?)\*(?!\w)/g,"<em>$1</em>"),i=i.replace(/\[([^\]]+)\]\(([^)]+)\)/g,'<a href="$2" target="_blank" rel="noopener">$1</a>'),i=i.replace(/((?:^|\n)\d+\. .+(?:\n\d+\. .+)*)/g,(d)=>{let x=d.trim().split(`
`),$="<ol>";for(let g of x)$+=`<li>${g.replace(/^\d+\.\s+/,"")}</li>`;return $+="</ol>",$}),i=i.replace(/((?:^|\n)- .+(?:\n- .+)*)/g,(d)=>{let x=d.trim().split(`
`),$="<ul>";for(let g of x)$+=`<li>${g.replace(/^-\s+/,"")}</li>`;return $+="</ul>",$}),i=i.replace(/\n\n/g,"</p><p>"),i=i.replace(/\n/g,"<br>"),!i.startsWith("<"))i=`<p>${i}</p>`;return i}esc(v){return v.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;")}}async function f(v,i,d,x){let $={Accept:"text/event-stream"},g;if(x)g=x;else $["Content-Type"]="application/json",g=JSON.stringify(i);let r=await fetch(v,{method:"POST",headers:$,body:g});if(!r.ok){let o=await r.text();d.onError?.(o||`HTTP ${r.status}`),d.onDone?.();return}let k=r.body.getReader(),T=new TextDecoder,c="",L="";while(!0){let{done:o,value:B}=await k.read();if(o)break;c+=T.decode(B,{stream:!0});let J=c.split(`
`);c=J.pop();for(let W of J){let G=W.trim();if(G.startsWith("event:")){L=G.slice(6).trim();continue}if(!G.startsWith("data:"))continue;let X=G.slice(5).trim();if(X==="[DONE]"){d.onDone?.();return}try{let Y=JSON.parse(X);switch(L){case"run_id":d.onRunId?.(Y.run_id);break;case"step":d.onStep?.(Y);break;case"result":d.onResult?.(Y);break;case"error":d.onError?.(Y.error||"Unknown error");break;case"delta":d.onDelta?.(Y.delta);break;case"tool_start":d.onToolStart?.(Y);break;case"tool_end":d.onToolEnd?.(Y);break;case"tool_render":d.onToolRender?.(Y);break;default:if(Y.delta)d.onDelta?.(Y.delta);break}}catch{}L=""}}d.onDone?.()}async function H0(v,i){let d=await fetch(v,{method:"GET",headers:{Accept:"text/event-stream"}});if(!d.ok){let k=await d.text();i.onError?.(k||`HTTP ${d.status}`),i.onDone?.();return}let x=d.body.getReader(),$=new TextDecoder,g="",r="";while(!0){let{done:k,value:T}=await x.read();if(k)break;g+=$.decode(T,{stream:!0});let c=g.split(`
`);g=c.pop();for(let L of c){let o=L.trim();if(o.startsWith("event:")){r=o.slice(6).trim();continue}if(!o.startsWith("data:"))continue;let B=o.slice(5).trim();if(B==="[DONE]"){i.onDone?.();return}try{let J=JSON.parse(B);switch(r){case"step":i.onStep?.(J);break;case"result":i.onResult?.(J);break;case"error":i.onError?.(J.error||"Unknown error");break;default:break}}catch{}r=""}}i.onDone?.()}async function h0(v,i,d,x,$){let g;if(i==="GET"||i==="DELETE"){let c=new URLSearchParams(d),L=c.toString()?`${v}?${c}`:v;g={method:i},v=L}else if(x&&$)g={method:i,body:$};else g={method:i,headers:{"Content-Type":"application/json"},body:JSON.stringify(d)};let r=await fetch(v,g),k=await r.text(),T;try{T=JSON.parse(k)}catch{T=k}return{status:r.status,data:T}}class g0 extends HTMLElement{meta;connectedCallback(){this.meta=JSON.parse(this.getAttribute("data-meta")||"{}"),this.render()}render(){let v=this.meta,i=this.attachShadow({mode:"open"});i.innerHTML=`
      <style>
        ${R}
        :host {
          display: block;
          padding: 2.5rem 1rem 3rem;
        }
        .layout {
          max-width: 960px;
          margin: 0 auto;
          width: 100%;
        }
        @media (min-width: 768px) {
          :host {
            padding: 2.5rem 2rem 3rem;
          }
        }

        /* Header */
        .header {
          margin-bottom: 1.5rem;
        }
        h1 {
          font-size: 1.3rem;
          font-weight: 700;
          color: var(--haira-text);
          display: flex;
          align-items: center;
          gap: 0.6rem;
          margin-bottom: 0.25rem;
        }
        .method-badge {
          font-size: 0.6rem;
          font-weight: 700;
          padding: 0.15rem 0.45rem;
          border-radius: 4px;
          color: #fff;
          letter-spacing: 0.02em;
          flex-shrink: 0;
        }
        .desc {
          font-size: 0.85rem;
          color: var(--haira-muted);
          line-height: 1.45;
        }
        .path {
          font-family: var(--haira-mono);
          font-size: 0.78rem;
          color: var(--haira-muted);
          opacity: 0.7;
          margin-top: 0.15rem;
        }

        /* Form card */
        .card {
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          padding: 1.25rem;
          transition: opacity 0.2s;
        }
        .card.disabled {
          opacity: 0.45;
          pointer-events: none;
        }

        /* Submit button */
        .submit-btn {
          width: 100%;
          padding: 0.65rem 1.5rem;
          border: none;
          background: var(--haira-accent);
          color: #0a0a0a;
          border-radius: var(--haira-radius-sm);
          font-size: 0.88rem;
          font-weight: 600;
          cursor: pointer;
          font-family: var(--haira-font);
          transition: all 0.15s;
          margin-top: 0.5rem;
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 0.4rem;
        }
        .submit-btn:hover:not(:disabled) {
          background: var(--haira-accent-light);
          box-shadow: 0 2px 16px rgba(232, 163, 23, 0.2);
        }
        .submit-btn:active:not(:disabled) {
          transform: scale(0.99);
        }
        .submit-btn:disabled {
          opacity: 0.5;
          cursor: not-allowed;
        }
        .spinner {
          display: inline-block;
          width: 14px;
          height: 14px;
          border: 2px solid #0a0a0a;
          border-top-color: transparent;
          border-radius: 50%;
          animation: spin 0.6s linear infinite;
        }
        @keyframes spin { to { transform: rotate(360deg); } }

        /* Pipeline + result area */
        .output-area {
          margin-top: 0.5rem;
        }
      </style>
      <div class="layout">
        <div class="header">
          <h1>
            ${this.esc(v.title||v.name)}
            <span class="method-badge" style="background:${D(v.method)}">${v.method}</span>
          </h1>
          ${v.description?`<p class="desc">${this.esc(v.description)}</p>`:""}
          <p class="path">${this.esc(v.path)}</p>
        </div>
        <div class="card" id="card">
          <form id="wf-form">
            <div id="fields"></div>
            <button type="submit" class="submit-btn" id="submit-btn">Run</button>
          </form>
        </div>
        <div class="output-area" id="output-area"></div>
      </div>
    `;let d=i.getElementById("fields");for(let L of v.params){let o=document.createElement("haira-field");o.setAttribute("name",L.Name),o.setAttribute("type",L.Type),d.appendChild(o)}let x=i.getElementById("output-area"),$=document.createElement("haira-pipeline");if(x.appendChild($),v.steps&&v.steps.length>0)$.setSteps(v.steps);let g=document.createElement("haira-result");x.appendChild(g);let r=i.getElementById("wf-form"),k=i.getElementById("submit-btn"),T=i.getElementById("card"),c=new URLSearchParams(window.location.search).get("run");if(c)this.loadRun(c,$,g,T,k);r.addEventListener("submit",async(L)=>{L.preventDefault(),k.disabled=!0,k.innerHTML='<span class="spinner"></span>Running...',T.classList.add("disabled"),g.hide();let o=d.querySelectorAll("haira-field"),B={},J,W=!1;for(let X of o){let{name:Y,value:E,type:b}=X.getValue();if(b==="file"&&E){if(W=!0,!J)J=new FormData;J.append(Y,E)}else if(E!==""&&E!==null){if(B[Y]=E,J)J.append(Y,String(E))}}let G=()=>{k.disabled=!1,k.textContent="Run",T.classList.remove("disabled"),history.replaceState(null,"",window.location.pathname)};if(v.steps&&v.steps.length>0){$.reset(),$.show();let X;if(W&&J){for(let[Y,E]of Object.entries(B))if(!J.has(Y))J.append(Y,String(E));X=J}await f(v.path,B,{onRunId:(Y)=>{history.replaceState(null,"",`${window.location.pathname}?run=${Y}`)},onStep:(Y)=>{$.updateStep(Y)},onResult:(Y)=>{g.show(Y,!1)},onError:(Y)=>{g.show({error:Y},!0)},onDone:()=>{$.finalize(),G()}},X)}else{try{let X=await h0(v.path,v.method,B,W,J);g.show(X.data,X.status>=400)}catch(X){g.show({error:X.message},!0)}G()}})}async loadRun(v,i,d,x,$){let g;try{let r=await fetch(`/_api/runs/${v}`);if(!r.ok)return;g=await r.json()}catch{return}i.show();for(let r of g.steps)i.updateStep(r);if(g.status==="completed"&&g.result)d.show(g.result,!1),i.finalize();else if(g.status==="failed"){if(g.error)d.show({error:g.error},!0);i.finalize()}else if(g.status==="running")$.disabled=!0,$.innerHTML='<span class="spinner"></span>Running...',x.classList.add("disabled"),await H0(`/_api/runs/stream/${v}`,{onStep:(r)=>{i.updateStep(r)},onResult:(r)=>{d.show(r,!1)},onError:(r)=>{d.show({error:r},!0)},onDone:()=>{i.finalize(),$.disabled=!1,$.textContent="Run",x.classList.remove("disabled"),history.replaceState(null,"",window.location.pathname)}})}esc(v){return v.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;")}}class x0 extends HTMLElement{connectedCallback(){let i=JSON.parse(this.getAttribute("data-meta")||"{}").workflows||[],d=this.attachShadow({mode:"open"});d.innerHTML=`
      <style>
        ${R}
        ${z}
        :host {
          display: flex;
          justify-content: center;
          padding: 2.5rem 1rem;
        }
        .container { max-width: 960px; width: 100%; }
        @media (min-width: 768px) {
          :host {
            padding: 2.5rem 2rem;
          }
        }
        h1 {
          font-size: 1.3rem;
          font-weight: 700;
          color: var(--haira-text);
          margin-bottom: 1.25rem;
        }
        .wf {
          display: flex;
          align-items: center;
          justify-content: space-between;
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          padding: 0.85rem 1rem;
          margin-bottom: 0.5rem;
          text-decoration: none;
          color: var(--haira-text);
          transition: all 0.15s;
          animation: fadeSlideUp 0.3s ease-out both;
        }
        .wf:hover {
          border-color: rgba(232, 163, 23, 0.3);
          background: var(--haira-bg-card-hover);
        }
        .wf-left {
          display: flex;
          align-items: center;
          gap: 0.6rem;
          min-width: 0;
        }
        .badge {
          font-size: 0.6rem;
          font-weight: 700;
          padding: 0.12rem 0.4rem;
          border-radius: 3px;
          color: #fff;
          flex-shrink: 0;
          letter-spacing: 0.02em;
        }
        .wf-info { min-width: 0; }
        .wf-name {
          font-weight: 600;
          font-size: 0.88rem;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
        }
        .wf-path {
          font-family: var(--haira-mono);
          font-size: 0.75rem;
          color: var(--haira-muted);
          margin-top: 0.1rem;
        }
        .wf-right {
          display: flex;
          align-items: center;
          flex-shrink: 0;
          margin-left: 0.75rem;
        }
        .type-pill {
          font-size: 0.65rem;
          font-weight: 600;
          padding: 0.12rem 0.5rem;
          border-radius: 10px;
          border: 1px solid;
          text-transform: lowercase;
        }
        .empty {
          text-align: center;
          padding: 3rem 1rem;
          animation: fadeIn 0.4s ease-out;
        }
        .empty-title {
          color: var(--haira-text-dim);
          font-size: 0.95rem;
          font-weight: 500;
          margin-bottom: 0.25rem;
        }
        .empty-sub {
          color: var(--haira-muted);
          font-size: 0.82rem;
        }

        /* Recent Runs */
        .section-title {
          font-size: 1rem;
          font-weight: 700;
          color: var(--haira-text);
          margin: 2rem 0 0.75rem;
        }
        .run {
          display: flex;
          align-items: center;
          gap: 0.6rem;
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius-sm);
          padding: 0.6rem 0.85rem;
          margin-bottom: 0.35rem;
          text-decoration: none;
          color: var(--haira-text);
          transition: all 0.15s;
          animation: fadeIn 0.25s ease-out both;
        }
        .run:hover {
          border-color: rgba(232, 163, 23, 0.3);
          background: var(--haira-bg-card-hover);
        }
        .run-dot {
          width: 7px;
          height: 7px;
          border-radius: 50%;
          flex-shrink: 0;
        }
        .run-dot.completed { background: var(--haira-success); }
        .run-dot.failed { background: var(--haira-error); }
        .run-dot.running {
          background: var(--haira-accent);
          animation: pulse 1.5s ease-in-out infinite;
        }
        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.4; }
        }
        .run-name {
          flex: 1;
          font-size: 0.82rem;
          font-weight: 500;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
          min-width: 0;
        }
        .run-name .run-id {
          font-family: var(--haira-mono);
          font-size: 0.7rem;
          color: var(--haira-muted);
          margin-left: 0.4rem;
        }
        .run-time {
          font-size: 0.72rem;
          font-family: var(--haira-mono);
          color: var(--haira-muted);
          flex-shrink: 0;
        }
        .run-status {
          font-size: 0.62rem;
          font-weight: 600;
          text-transform: uppercase;
          letter-spacing: 0.03em;
          flex-shrink: 0;
          padding: 0.08rem 0.35rem;
          border-radius: 3px;
        }
        .run-status.completed {
          color: var(--haira-success);
          background: rgba(34, 197, 94, 0.1);
        }
        .run-status.failed {
          color: var(--haira-error);
          background: rgba(239, 68, 68, 0.1);
        }
        .run-status.running {
          color: var(--haira-accent);
          background: rgba(232, 163, 23, 0.1);
        }

        /* Recent Chats */
        .chat {
          display: flex;
          align-items: center;
          gap: 0.6rem;
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius-sm);
          padding: 0.6rem 0.85rem;
          margin-bottom: 0.35rem;
          text-decoration: none;
          color: var(--haira-text);
          transition: all 0.15s;
          animation: fadeIn 0.25s ease-out both;
        }
        .chat:hover {
          border-color: rgba(232, 163, 23, 0.3);
          background: var(--haira-bg-card-hover);
        }
        .chat-icon {
          color: var(--haira-accent);
          display: flex;
          flex-shrink: 0;
          opacity: 0.6;
        }
        .chat-title {
          flex: 1;
          font-size: 0.82rem;
          font-weight: 500;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
          min-width: 0;
        }
        .chat-wf {
          font-family: var(--haira-mono);
          font-size: 0.7rem;
          color: var(--haira-muted);
          margin-left: 0.4rem;
        }
        .chat-time {
          font-size: 0.72rem;
          font-family: var(--haira-mono);
          color: var(--haira-muted);
          flex-shrink: 0;
        }
        .chat-count {
          font-size: 0.62rem;
          color: var(--haira-muted);
          flex-shrink: 0;
          padding: 0.08rem 0.35rem;
          border-radius: 3px;
          background: var(--haira-bg-elevated);
        }
      </style>
      <div class="container">
        <h1>Workflows</h1>
        ${i.length===0?`<div class="empty">
              <div class="empty-title">No workflows registered</div>
              <div class="empty-sub">Define a workflow in your .haira file to get started</div>
            </div>`:i.map((x,$)=>`
            <a class="wf" href="/_ui${x.path}" style="animation-delay:${$*50}ms">
              <div class="wf-left">
                <span class="badge" style="background:${D(x.method)}">${x.method}</span>
                <div class="wf-info">
                  <div class="wf-name">${this.esc(x.title||x.name)}</div>
                  <div class="wf-path">${this.esc(x.path)}</div>
                </div>
              </div>
              <div class="wf-right">
                <span class="type-pill" style="color:${n(x.uiType)};border-color:${n(x.uiType)}30">${x.uiType}</span>
              </div>
            </a>
          `).join("")}
        <div id="chats-section"></div>
        <div id="runs-section"></div>
      </div>
    `,this.loadChats(d),this.loadRuns(d)}async loadChats(v){let i=v.getElementById("chats-section");if(!i)return;try{let d=await fetch("/_api/chats");if(!d.ok)return;let x=await d.json();if(!x||x.length===0)return;let $='<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z"/></svg>';i.innerHTML=`
        <h2 class="section-title">Recent Chats</h2>
        ${x.map((g,r)=>`
          <a class="chat" href="/_ui${this.esc(g.workflow_path)}?session=${this.esc(g.id)}" style="animation-delay:${r*30}ms">
            <span class="chat-icon">${$}</span>
            <span class="chat-title">${this.esc(g.title||"New chat")}<span class="chat-wf">${this.esc(g.workflow_name)}</span></span>
            <span class="chat-time">${this.relativeTime(g.updated_at)}</span>
            <span class="chat-count">${g.message_count} msg</span>
          </a>
        `).join("")}
      `}catch{}}async loadRuns(v){let i=v.getElementById("runs-section");if(!i)return;try{let d=await fetch("/_api/runs");if(!d.ok)return;let x=await d.json();if(!x||x.length===0)return;i.innerHTML=`
        <h2 class="section-title">Recent Runs</h2>
        ${x.map(($,g)=>`
          <a class="run" href="/_ui${this.esc($.workflow_path)}?run=${this.esc($.id)}" style="animation-delay:${g*30}ms">
            <span class="run-dot ${$.status}"></span>
            <span class="run-name">${this.esc($.workflow_name)}<span class="run-id">${this.shortId($.id)}</span></span>
            <span class="run-time">${this.relativeTime($.started_at)}</span>
            <span class="run-status ${$.status}">${$.status}</span>
          </a>
        `).join("")}
      `}catch{}}relativeTime(v){let i=Date.now(),d=new Date(v).getTime(),x=i-d,$=Math.floor(x/1000);if($<60)return"just now";let g=Math.floor($/60);if(g<60)return`${g}m ago`;let r=Math.floor(g/60);if(r<24)return`${r}h ago`;return`${Math.floor(r/24)}d ago`}shortId(v){let i=v.split("_");if(i.length>=4)return i.slice(2).join("_");return v}esc(v){return v.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;")}}var q0='<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21.44 11.05l-9.19 9.19a6 6 0 01-8.49-8.49l9.19-9.19a4 4 0 015.66 5.66l-9.2 9.19a2 2 0 01-2.83-2.83l8.49-8.48"/></svg>',u0='<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>',l0='<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>',C0='<svg width="12" height="12" viewBox="0 0 16 16" fill="none"><path d="M5 5L11 11M11 5L5 11" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>',E0='<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>';var m0='<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 18 9 12 15 6"/></svg>',n0='<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>',p0='<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z"/></svg>',s0='<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/></svg>',e0='<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2"/><line x1="9" y1="3" x2="9" y2="21"/></svg>';class $0 extends HTMLElement{meta;sessionId="";attachedFile=null;connectedCallback(){this.meta=JSON.parse(this.getAttribute("data-meta")||"{}");let v=new URL(window.location.href),i=v.searchParams.get("session");if(i)this.sessionId=i;else this.sessionId=crypto.randomUUID(),v.searchParams.set("session",this.sessionId),window.history.replaceState({},"",v.toString());this.render()}render(){let v=this.meta,i=this.shadowRoot||this.attachShadow({mode:"open"});i.innerHTML=`
      <style>
        ${R}
        ${z}
        :host {
          display: flex;
          flex-direction: row;
          flex: 1;
          overflow: hidden;
          position: relative;
        }

        /* ---- Session sidebar ---- */
        .sidebar {
          width: 240px;
          flex-shrink: 0;
          display: flex;
          flex-direction: column;
          border-right: 1px solid var(--haira-border);
          background: var(--haira-bg);
          overflow: hidden;
          transition: width 0.2s, opacity 0.2s;
        }
        .sidebar.collapsed {
          width: 0;
          opacity: 0;
          pointer-events: none;
        }
        .sidebar-header {
          display: flex;
          align-items: center;
          gap: 0.4rem;
          padding: 0.55rem 0.65rem;
          border-bottom: 1px solid var(--haira-border);
          flex-shrink: 0;
        }
        .sidebar-title {
          font-size: 0.75rem;
          font-weight: 600;
          color: var(--haira-text-dim);
          flex: 1;
        }
        .sidebar-btn {
          background: none;
          border: none;
          color: var(--haira-muted);
          cursor: pointer;
          display: flex;
          align-items: center;
          justify-content: center;
          padding: 0.25rem;
          border-radius: 4px;
          transition: all 0.15s;
        }
        .sidebar-btn:hover {
          color: var(--haira-accent);
          background: var(--haira-accent-dim);
        }
        .sidebar-list {
          flex: 1;
          overflow-y: auto;
          padding: 0.35rem;
          display: flex;
          flex-direction: column;
          gap: 1px;
          ${A}
        }
        .session-item {
          display: flex;
          align-items: center;
          gap: 0.4rem;
          padding: 0.45rem 0.5rem;
          border-radius: 6px;
          cursor: pointer;
          transition: all 0.12s;
          text-decoration: none;
          color: var(--haira-text-dim);
          font-size: 0.78rem;
          line-height: 1.35;
          position: relative;
        }
        .session-item:hover {
          background: var(--haira-bg-card);
          color: var(--haira-text);
        }
        .session-item.active {
          background: var(--haira-accent-dim);
          color: var(--haira-accent);
        }
        .session-icon {
          display: flex;
          flex-shrink: 0;
          opacity: 0.5;
        }
        .session-item.active .session-icon {
          opacity: 1;
        }
        .session-title {
          flex: 1;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
          min-width: 0;
        }
        .session-delete {
          display: none;
          background: none;
          border: none;
          color: var(--haira-muted);
          cursor: pointer;
          padding: 0.15rem;
          border-radius: 3px;
          flex-shrink: 0;
          align-items: center;
          justify-content: center;
        }
        .session-item:hover .session-delete {
          display: flex;
        }
        .session-delete:hover {
          color: var(--haira-error);
          background: rgba(239, 68, 68, 0.1);
        }
        .sidebar-empty {
          padding: 1rem;
          text-align: center;
          font-size: 0.75rem;
          color: var(--haira-muted);
          opacity: 0.5;
        }

        /* Sidebar toggle (in chat-main when sidebar collapsed) */
        .sidebar-toggle {
          position: absolute;
          top: 0.5rem;
          left: 0.5rem;
          z-index: 10;
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          color: var(--haira-muted);
          cursor: pointer;
          display: none;
          align-items: center;
          justify-content: center;
          padding: 0.35rem;
          border-radius: 6px;
          transition: all 0.15s;
        }
        .sidebar-toggle.visible {
          display: flex;
        }
        .sidebar-toggle:hover {
          color: var(--haira-accent);
          border-color: var(--haira-accent);
          background: var(--haira-accent-dim);
        }

        /* ---- Chat main column ---- */
        .chat-main {
          flex: 1;
          display: flex;
          flex-direction: column;
          overflow: hidden;
          position: relative;
          min-width: 0;
          height: 100%;
        }

        /* Welcome screen */
        .welcome {
          flex: 1;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          gap: 1rem;
          padding: 2rem;
          opacity: 1;
          transition: opacity 0.3s;
        }
        .welcome.hidden {
          display: none;
        }
        .welcome-icon {
          opacity: 0.15;
        }
        .welcome-icon img {
          width: 56px;
          height: 56px;
          object-fit: contain;
          opacity: 1;
        }
        .welcome h2 {
          font-size: 1.1rem;
          font-weight: 600;
          color: var(--haira-text);
        }
        .welcome p {
          font-size: 0.85rem;
          color: var(--haira-muted);
          text-align: center;
          max-width: 420px;
          line-height: 1.5;
        }
        .suggestions {
          display: flex;
          flex-wrap: wrap;
          gap: 0.5rem;
          justify-content: center;
          margin-top: 0.5rem;
          max-width: 540px;
        }
        .suggestion {
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          color: var(--haira-text-dim);
          padding: 0.45rem 0.85rem;
          border-radius: 20px;
          font-size: 0.78rem;
          font-family: var(--haira-font);
          cursor: pointer;
          transition: all 0.15s;
        }
        .suggestion:hover {
          border-color: var(--haira-accent);
          color: var(--haira-accent);
          background: var(--haira-accent-dim);
        }

        /* Messages area */
        .messages {
          flex: 1;
          min-height: 0;
          overflow-y: auto;
          display: none;
          flex-direction: column;
          ${A}
        }
        .messages.active {
          display: flex;
        }
        .messages-inner {
          max-width: 768px;
          width: 100%;
          margin: 0 auto;
          padding: 1.5rem 1.25rem;
          display: flex;
          flex-direction: column;
          gap: 0.75rem;
        }

        /* Typing indicator (inside messages-inner) */
        .typing {
          display: none;
          padding: 0.25rem 0;
          align-items: center;
          gap: 0.4rem;
          font-size: 0.75rem;
          color: var(--haira-muted);
          margin-left: 2.25rem;
        }
        .typing.visible { display: flex; }
        .typing-dots {
          display: flex;
          gap: 0.2rem;
          align-items: center;
        }
        .typing-dot {
          display: inline-block;
          width: 5px;
          height: 5px;
          border-radius: 50%;
          background: var(--haira-accent);
          animation: bounce 1.4s ease-in-out infinite;
        }
        .typing-dot:nth-child(2) { animation-delay: 0.2s; }
        .typing-dot:nth-child(3) { animation-delay: 0.4s; }

        /* Drop overlay */
        .drop-overlay {
          display: none;
          position: absolute;
          inset: 0;
          background: rgba(9, 9, 11, 0.85);
          z-index: 200;
          align-items: center;
          justify-content: center;
          flex-direction: column;
          gap: 0.75rem;
          border: 2px dashed var(--haira-accent);
          border-radius: var(--haira-radius);
          margin: 0.5rem;
        }
        .drop-overlay.visible {
          display: flex;
        }
        .drop-overlay-icon {
          color: var(--haira-accent);
          opacity: 0.7;
        }
        .drop-overlay-text {
          color: var(--haira-accent);
          font-size: 0.9rem;
          font-weight: 600;
        }

        /* Input area — pinned to bottom by flex layout */
        .input-area {
          padding: 0.75rem 1rem 1rem;
          flex-shrink: 0;
          background: var(--haira-bg);
          border-top: 1px solid var(--haira-border);
        }
        .input-card {
          display: flex;
          flex-direction: column;
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          transition: border-color 0.15s;
          max-width: 768px;
          margin: 0 auto;
        }
        .input-card:focus-within {
          border-color: var(--haira-border-focus);
        }

        /* File chip */
        .file-chip {
          display: none;
          align-items: center;
          gap: 0.4rem;
          padding: 0.4rem 0.6rem 0;
          margin: 0 0.5rem;
        }
        .file-chip.visible { display: flex; }
        .file-chip-inner {
          display: flex;
          align-items: center;
          gap: 0.35rem;
          background: var(--haira-bg-elevated);
          border: 1px solid var(--haira-border);
          border-radius: 6px;
          padding: 0.25rem 0.5rem;
          font-size: 0.75rem;
          color: var(--haira-text-dim);
        }
        .file-chip-icon { color: var(--haira-accent); display: flex; }
        .file-chip-name {
          max-width: 200px;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
        .file-chip-size {
          color: var(--haira-muted);
          font-size: 0.7rem;
        }
        .file-chip-remove {
          background: none;
          border: none;
          color: var(--haira-muted);
          cursor: pointer;
          display: flex;
          padding: 0.1rem;
          border-radius: 3px;
          transition: all 0.15s;
        }
        .file-chip-remove:hover {
          color: var(--haira-error);
          background: rgba(239, 68, 68, 0.1);
        }

        .input-row {
          display: flex;
          align-items: flex-end;
          gap: 0.35rem;
          padding: 0.35rem;
        }
        .attach-btn {
          background: none;
          border: none;
          color: var(--haira-muted);
          cursor: pointer;
          display: flex;
          align-items: center;
          justify-content: center;
          padding: 0.45rem;
          border-radius: 6px;
          transition: all 0.15s;
          flex-shrink: 0;
        }
        .attach-btn:hover {
          color: var(--haira-accent);
          background: var(--haira-accent-dim);
        }
        textarea {
          flex: 1;
          background: transparent;
          border: none;
          color: var(--haira-text);
          padding: 0.5rem 0.35rem;
          font-size: 0.9rem;
          font-family: var(--haira-font);
          resize: none;
          min-height: 44px;
          max-height: 200px;
          outline: none;
          line-height: 1.5;
        }
        textarea::placeholder {
          color: var(--haira-muted);
        }
        .send-btn {
          background: var(--haira-accent);
          color: #1a0e04;
          border: none;
          width: 34px;
          height: 34px;
          border-radius: 8px;
          cursor: pointer;
          display: flex;
          align-items: center;
          justify-content: center;
          transition: all 0.15s;
          flex-shrink: 0;
        }
        .send-btn:hover {
          background: var(--haira-accent-light);
          box-shadow: 0 2px 12px rgba(232, 163, 23, 0.25);
        }
        .send-btn:disabled {
          opacity: 0.35;
          cursor: not-allowed;
          box-shadow: none;
        }
        .input-hint {
          text-align: center;
          font-size: 0.68rem;
          color: var(--haira-muted);
          opacity: 0.5;
          padding-top: 0.35rem;
        }

        /* ---- Activity panel (floating overlay) ---- */
        .activity-panel {
          position: absolute;
          right: 0;
          top: 0;
          bottom: 0;
          width: 280px;
          z-index: 50;
          display: flex;
          flex-direction: column;
          border-left: 1px solid var(--haira-border);
          background: var(--haira-bg);
          overflow: hidden;
          box-shadow: -4px 0 24px rgba(0, 0, 0, 0.25);
        }
        .activity-panel.collapsed {
          display: none;
        }
        .panel-header {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          padding: 0.55rem 0.75rem;
          border-bottom: 1px solid var(--haira-border);
          flex-shrink: 0;
        }
        .panel-header-icon {
          display: flex;
          color: var(--haira-muted);
        }
        .panel-title {
          font-size: 0.78rem;
          font-weight: 600;
          color: var(--haira-text-dim);
          flex: 1;
        }
        .panel-count {
          font-size: 0.68rem;
          color: var(--haira-muted);
          font-family: var(--haira-mono);
        }
        .panel-close {
          background: none;
          border: none;
          color: var(--haira-muted);
          cursor: pointer;
          display: flex;
          align-items: center;
          justify-content: center;
          padding: 0.2rem;
          border-radius: 4px;
          transition: all 0.15s;
        }
        .panel-close:hover {
          color: var(--haira-text);
          background: var(--haira-bg-elevated);
        }
        .panel-body {
          flex: 1;
          overflow-y: auto;
          padding: 0.5rem;
          display: flex;
          flex-direction: column;
          gap: 0.4rem;
          ${A}
        }
        .panel-empty {
          display: flex;
          align-items: center;
          justify-content: center;
          flex: 1;
          font-size: 0.75rem;
          color: var(--haira-muted);
          opacity: 0.5;
        }

        /* Toggle button (floating in chat-main) */
        .activity-toggle {
          position: absolute;
          top: 0.5rem;
          right: 0.5rem;
          z-index: 10;
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          color: var(--haira-muted);
          cursor: pointer;
          display: flex;
          align-items: center;
          justify-content: center;
          padding: 0.35rem;
          border-radius: 6px;
          transition: all 0.15s;
          gap: 0.3rem;
        }
        .activity-toggle:hover {
          color: var(--haira-accent);
          border-color: var(--haira-accent);
          background: var(--haira-accent-dim);
        }
        .activity-toggle .badge {
          display: none;
          min-width: 16px;
          height: 16px;
          padding: 0 4px;
          border-radius: 8px;
          background: var(--haira-accent);
          color: #1a0e04;
          font-size: 0.62rem;
          font-weight: 700;
          line-height: 16px;
          text-align: center;
        }
        .activity-toggle .badge.visible {
          display: inline-block;
        }

        /* Mobile */
        @media (max-width: 640px) {
          .sidebar {
            position: absolute;
            left: 0;
            top: 0;
            bottom: 0;
            z-index: 100;
            box-shadow: 4px 0 24px rgba(0, 0, 0, 0.3);
          }
          .messages-inner {
            padding: 1rem 0.75rem;
          }
          .input-area {
            padding: 0.5rem 0.5rem 0.75rem;
            padding-bottom: max(0.75rem, env(safe-area-inset-bottom));
          }
          .welcome {
            padding: 1.5rem 1rem;
          }
          .suggestions {
            max-width: 100%;
          }
        }
      </style>

      <div class="sidebar" id="sidebar">
        <div class="sidebar-header">
          <span class="sidebar-title">Chats</span>
          <button class="sidebar-btn" id="new-chat-btn" title="New chat">${n0}</button>
          <button class="sidebar-btn" id="sidebar-close-btn" title="Close sidebar">${m0}</button>
        </div>
        <div class="sidebar-list" id="sidebar-list">
          <div class="sidebar-empty" id="sidebar-empty">No chats yet</div>
        </div>
      </div>

      <div class="chat-main">
        <button class="sidebar-toggle" id="sidebar-open-btn" title="Show chats">${e0}</button>

        <div class="welcome" id="welcome">
          <span class="welcome-icon">${v.logo?`<img src="${this.esc(v.logo)}" alt="">`:O.replace(/width="22" height="22"/,'width="56" height="56"')}</span>
          <h2>${this.esc(v.title||v.name||"Chat")}</h2>
          ${v.description?`<p>${this.esc(v.description)}</p>`:""}
          <div class="suggestions" id="suggestions"></div>
        </div>

        <div class="messages" id="messages">
          <div class="messages-inner" id="messages-inner">
            <div class="typing" id="typing">
              <div class="typing-dots">
                <span class="typing-dot"></span>
                <span class="typing-dot"></span>
                <span class="typing-dot"></span>
              </div>
              <span>Thinking...</span>
            </div>
          </div>
        </div>

        <div class="drop-overlay" id="drop-overlay">
          <span class="drop-overlay-icon">${q0}</span>
          <span class="drop-overlay-text">Drop file to attach</span>
        </div>

        <button class="activity-toggle" id="activity-toggle" title="Toggle activity panel">
          ${E0}
          <span class="badge" id="toggle-badge">0</span>
        </button>

        <div class="input-area">
          <div class="input-card" id="input-card">
            <div class="file-chip" id="file-chip">
              <div class="file-chip-inner">
                <span class="file-chip-icon">${l0}</span>
                <span class="file-chip-name" id="file-name"></span>
                <span class="file-chip-size" id="file-size"></span>
                <button class="file-chip-remove" id="file-remove" title="Remove file">${C0}</button>
              </div>
            </div>
            <div class="input-row">
              ${v.hasFile?`<button class="attach-btn" id="attach-btn" title="Attach file">${q0}</button>`:""}
              <textarea id="chat-input" placeholder="${v.hasFile?"Type a message or drop a file...":"Type a message..."}" rows="1"></textarea>
              <button class="send-btn" id="send-btn" title="Send">${u0}</button>
            </div>
          </div>
          ${v.hasFile?'<input type="file" id="file-input" style="display:none" />':""}
          <div class="input-hint">Enter to send, Shift+Enter for new line</div>
        </div>
      </div>

      <div class="activity-panel collapsed" id="activity-panel">
        <div class="panel-header">
          <span class="panel-header-icon">${E0}</span>
          <span class="panel-title">Activity</span>
          <span class="panel-count" id="panel-count"></span>
          <button class="panel-close" id="panel-close" title="Close panel">${C0}</button>
        </div>
        <div class="panel-body" id="panel-body">
          <div class="panel-empty" id="panel-empty">No activity yet</div>
        </div>
      </div>
    `;let d=i.getElementById("messages"),x=i.getElementById("messages-inner"),$=i.getElementById("welcome"),g=i.getElementById("chat-input"),r=i.getElementById("send-btn"),k=i.getElementById("file-chip"),T=i.getElementById("file-name"),c=i.getElementById("file-size"),L=i.getElementById("file-remove"),o=i.getElementById("typing"),B=i.getElementById("drop-overlay"),J=i.getElementById("suggestions"),W=v.hasFile?i.getElementById("attach-btn"):null,G=v.hasFile?i.getElementById("file-input"):null,X=i.getElementById("activity-panel"),Y=i.getElementById("panel-body"),E=i.getElementById("panel-empty"),b=i.getElementById("panel-count"),F0=i.getElementById("panel-close"),y0=i.getElementById("activity-toggle"),t=i.getElementById("toggle-badge"),M0=i.getElementById("sidebar"),a=i.getElementById("sidebar-list"),Z0=i.getElementById("sidebar-empty"),w0=i.getElementById("new-chat-btn"),O0=i.getElementById("sidebar-close-btn"),J0=i.getElementById("sidebar-open-btn"),F=!1,N=0,u=0;function l(Q){F=Q!==void 0?Q:!F,X.classList.toggle("collapsed",!F)}function U0(){if(N>0)t.textContent=String(N),t.classList.add("visible");else t.classList.remove("visible");b.textContent=u>0?String(u):""}y0.addEventListener("click",()=>l()),F0.addEventListener("click",()=>l(!1));let y=!0,j0=(Q)=>{y=Q!==void 0?Q:!y,M0.classList.toggle("collapsed",!y),J0.classList.toggle("visible",!y)};O0.addEventListener("click",()=>j0(!1)),J0.addEventListener("click",()=>j0(!0));let K=this,m=async()=>{try{let Q=await fetch(`/_api/chats?workflow=${encodeURIComponent(v.path)}`);if(!Q.ok)return;let j=await Q.json();if(!j||j.length===0){Z0.style.display="",a.querySelectorAll(".session-item").forEach((Z)=>Z.remove());return}Z0.style.display="none",a.querySelectorAll(".session-item").forEach((Z)=>Z.remove());for(let Z of j){let U=document.createElement("div");U.className=`session-item${Z.id===K.sessionId?" active":""}`,U.innerHTML=`
            <span class="session-icon">${p0}</span>
            <span class="session-title">${K.esc(Z.title||"New chat")}</span>
            <button class="session-delete" title="Delete">${s0}</button>
          `,U.addEventListener("click",(h)=>{if(h.target.closest(".session-delete"))return;K.switchSession(Z.id)}),U.querySelector(".session-delete").addEventListener("click",async(h)=>{if(h.stopPropagation(),await fetch(`/_api/chats/${Z.id}`,{method:"DELETE"}),Z.id===K.sessionId)K.startNewChat();m()}),a.appendChild(U)}}catch{}};w0.addEventListener("click",()=>{K.startNewChat()});let D0=this.getSuggestions();for(let Q of D0){let j=document.createElement("button");j.className="suggestion",j.textContent=Q,j.addEventListener("click",()=>{g.value=Q,M()}),J.appendChild(j)}if(W&&G){W.addEventListener("click",()=>G.click()),G.addEventListener("change",()=>{if(G.files&&G.files[0])this.setFile(G.files[0],k,T,c)}),L.addEventListener("click",()=>{this.clearFile(k,G)});let Q=this,j=0;i.addEventListener("dragenter",(Z)=>{Z.preventDefault(),j++,B.classList.add("visible")}),i.addEventListener("dragleave",(Z)=>{if(Z.preventDefault(),j--,j<=0)j=0,B.classList.remove("visible")}),i.addEventListener("dragover",(Z)=>{Z.preventDefault()}),i.addEventListener("drop",(Z)=>{Z.preventDefault(),j=0,B.classList.remove("visible");let U=Z.dataTransfer;if(U&&U.files&&U.files[0])Q.setFile(U.files[0],k,T,c)})}g.addEventListener("input",()=>{g.style.height="auto",g.style.height=`${Math.min(g.scrollHeight,200)}px`}),g.addEventListener("keydown",(Q)=>{if(Q.key==="Enter"&&!Q.shiftKey)Q.preventDefault(),M()}),r.addEventListener("click",M),i.addEventListener("haira-chat-input",(Q)=>{let j=Q.detail?.text;if(j&&!r.disabled)g.value=j,M()}),requestAnimationFrame(()=>g.focus());let f0=v.avatar||"H";function P(Q,j,Z){let U=document.createElement("haira-message");if(U.setAttribute("role",Q),U.setAttribute("content",j),Z)U.setAttribute("file",Z);if(Q==="assistant")U.setAttribute("avatar",f0);return x.insertBefore(U,o),d.scrollTop=d.scrollHeight,U}function S0(){x.querySelectorAll("haira-message, haira-tool-card, haira-ui-renderer").forEach((j)=>j.remove())}async function M(){let Q=g.value.trim();if(!Q&&!K.attachedFile)return;g.value="",g.style.height="auto",$.classList.add("hidden"),d.classList.add("active");let j=K.attachedFile?K.attachedFile.name:void 0;P("user",Q,j),r.disabled=!0,o.classList.add("visible"),d.scrollTop=d.scrollHeight;let Z=v.chatParam||"message",U,h={};if(h[Z]=Q,h.session_id=K.sessionId,K.attachedFile){let V=v.fileParam||"file_path";U=new FormData,U.append(V,K.attachedFile),U.append(Z,Q),U.append("session_id",K.sessionId)}if(G)K.clearFile(k,G);let q=new Map,C=null,_="";await f(v.path,h,{onToolStart:(V)=>{o.classList.remove("visible");let H=document.createElement("haira-tool-card");if(E.style.display="none",Y.appendChild(H),H.setTool(V.tool),q.set(V.tool,H),Y.scrollTop=Y.scrollHeight,N++,u++,U0(),!F)l(!0)},onToolRender:(V)=>{let H=document.createElement("haira-ui-renderer");x.insertBefore(H,o),requestAnimationFrame(()=>H.render(V)),d.scrollTop=d.scrollHeight},onToolEnd:(V)=>{let H=q.get(V.tool);if(H)H.complete(V.ok!==!1),q.delete(V.tool);o.classList.add("visible"),N=Math.max(0,N-1),U0()},onDelta:(V)=>{if(o.classList.remove("visible"),!C)C=P("assistant","");_+=V,C.updateContent(_),d.scrollTop=d.scrollHeight},onError:(V)=>{if(o.classList.remove("visible"),!C)C=P("assistant","");C.updateContent(`Error: ${V}`),r.disabled=!1,g.focus()},onDone:()=>{if(o.classList.remove("visible"),!C&&_==="")C=P("assistant",""),C.updateContent("No response received. Please check the server logs.");r.disabled=!1,g.focus(),m()}},U)}(async(Q)=>{try{let j=await fetch(`/_api/chats/${Q}`);if(!j.ok)return;let Z=await j.json();if(!Z.messages||Z.messages.length===0)return;$.classList.add("hidden"),d.classList.add("active"),S0();let U=Z.messages;for(let h=0;h<U.length;h++){let q=U[h],C=P(q.role,q.content);if(q.role==="assistant")C.updateContent(q.content);if(q.ui_events&&q.ui_events.length>0){let _=U.slice(h+1).some((V)=>V.role==="user");for(let V of q.ui_events){let H=document.createElement("haira-ui-renderer");if(_)H.setAttribute("data-restored","true");x.insertBefore(H,o),requestAnimationFrame(()=>H.render(V))}}}d.scrollTop=d.scrollHeight}catch{}})(this.sessionId),m()}switchSession(v){this.sessionId=v;let i=new URL(window.location.href);if(i.searchParams.set("session",v),window.history.pushState({},"",i.toString()),this.shadowRoot)this.shadowRoot.innerHTML="";this.render()}startNewChat(){let v=crypto.randomUUID();this.switchSession(v)}setFile(v,i,d,x){this.attachedFile=v,d.textContent=v.name,x.textContent=this.formatSize(v.size),i.classList.add("visible")}clearFile(v,i){this.attachedFile=null,v.classList.remove("visible"),i.value=""}formatSize(v){if(v<1024)return`${v} B`;if(v<1048576)return`${(v/1024).toFixed(1)} KB`;return`${(v/1048576).toFixed(1)} MB`}getSuggestions(){if(this.meta.suggestions&&this.meta.suggestions.length>0)return this.meta.suggestions;return["What can you help me with?","Hello!"]}esc(v){return v.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;")}}var v1='<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-dasharray="28 10" style="animation:spin 0.7s linear infinite;transform-origin:center"/></svg>',i1='<svg width="13" height="13" viewBox="0 0 16 16" fill="none"><path d="M4.5 8.5L7 11L11.5 5.5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>',d1='<svg width="13" height="13" viewBox="0 0 16 16" fill="none"><path d="M5 5L11 11M11 5L5 11" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>';function g1(v){return v.replace(/_/g," ").replace(/\b\w/g,(i)=>i.toUpperCase())}class r0 extends HTMLElement{startTime=0;connectedCallback(){this.startTime=Date.now();let v=this.attachShadow({mode:"open"});v.innerHTML=`
      <style>
        ${R}
        ${z}
        :host {
          display: block;
          animation: fadeSlideUp 0.2s ease-out;
        }
        .card {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          padding: 0.5rem 0.75rem;
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: 8px;
        }
        .icon {
          display: flex;
          align-items: center;
          justify-content: center;
          width: 24px;
          height: 24px;
          border-radius: 6px;
          flex-shrink: 0;
        }
        .icon.running {
          background: rgba(232, 163, 23, 0.1);
          color: var(--haira-accent);
        }
        .icon.done {
          background: rgba(34, 197, 94, 0.1);
          color: var(--haira-success);
        }
        .icon.failed {
          background: rgba(239, 68, 68, 0.1);
          color: var(--haira-error);
        }
        .info {
          flex: 1;
          min-width: 0;
        }
        .tool-name {
          font-size: 0.78rem;
          font-weight: 600;
          color: var(--haira-text);
        }
        .tool-status {
          font-size: 0.7rem;
          color: var(--haira-muted);
          display: flex;
          align-items: center;
          gap: 0.3rem;
        }
        .duration {
          font-family: var(--haira-mono);
          font-size: 0.68rem;
          color: var(--haira-muted);
          flex-shrink: 0;
        }
        @keyframes spin {
          to { transform: rotate(360deg); }
        }
      </style>
      <div class="card" id="card">
        <div class="icon running" id="icon">${v1}</div>
        <div class="info">
          <div class="tool-name" id="name"></div>
          <div class="tool-status" id="status">Running...</div>
        </div>
        <span class="duration" id="duration"></span>
      </div>
    `}setTool(v){let i=this.shadowRoot?.getElementById("name");if(i)i.textContent=g1(v)}complete(v){let i=this.shadowRoot?.getElementById("icon"),d=this.shadowRoot?.getElementById("status"),x=this.shadowRoot?.getElementById("duration"),g=((Date.now()-this.startTime)/1000).toFixed(1);if(i)i.className=`icon ${v?"done":"failed"}`,i.innerHTML=v?i1:d1;if(d)d.textContent=v?"Completed":"Failed";if(x)x.textContent=`${g}s`}}var S={success:'<svg width="16" height="16" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="7" stroke="currentColor" stroke-width="1.5"/><path d="M5 8.5L7 10.5L11 5.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>',error:'<svg width="16" height="16" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="7" stroke="currentColor" stroke-width="1.5"/><path d="M5.5 5.5L10.5 10.5M10.5 5.5L5.5 10.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>',warning:'<svg width="16" height="16" viewBox="0 0 16 16" fill="none"><path d="M8 2L14.5 13H1.5L8 2Z" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"/><path d="M8 6.5V9" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><circle cx="8" cy="11" r="0.75" fill="currentColor"/></svg>',info:'<svg width="16" height="16" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="7" stroke="currentColor" stroke-width="1.5"/><path d="M8 7V11" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><circle cx="8" cy="5" r="0.75" fill="currentColor"/></svg>'},I0={success:"var(--haira-success)",error:"var(--haira-error)",warning:"var(--haira-warn)",info:"var(--haira-info)"};class k0 extends HTMLElement{connectedCallback(){this.attachShadow({mode:"open"}),this.shadowRoot.innerHTML=`
      <style>
        ${R}
        ${z}
        :host {
          display: block;
          animation: fadeSlideUp 0.25s ease-out;
        }
        .card {
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          overflow: hidden;
        }
        .header {
          display: flex;
          align-items: center;
          gap: 0.4rem;
          padding: 0.45rem 0.75rem;
        }
        .icon { display: flex; align-items: center; flex-shrink: 0; }
        .icon svg { width: 14px; height: 14px; }
        .title { font-size: 0.78rem; font-weight: 600; }
        .message {
          font-size: 0.75rem;
          color: var(--haira-text-dim);
          padding: 0 0.75rem 0.45rem 2rem;
          line-height: 1.4;
        }
        .sections {
          border-top: 1px solid var(--haira-border);
        }
        .section {
          padding: 0.4rem 0.75rem;
          border-bottom: 1px solid var(--haira-border);
        }
        .section:last-child { border-bottom: none; }
        .section-label {
          font-size: 0.68rem;
          font-weight: 600;
          color: var(--haira-muted);
          text-transform: uppercase;
          letter-spacing: 0.04em;
          margin-bottom: 0.2rem;
        }
        .section-content {
          font-size: 0.75rem;
          color: var(--haira-text-dim);
          line-height: 1.4;
          white-space: pre-wrap;
        }
        .section-content.code {
          font-family: var(--haira-mono);
          font-size: 0.72rem;
          background: var(--haira-bg);
          padding: 0.35rem 0.6rem;
          border-radius: var(--haira-radius-sm);
          overflow-x: auto;
        }
        /* Inline variant: title + message on one line, no sections */
        .card.inline .header {
          padding: 0.35rem 0.65rem;
        }
        .card.inline .message {
          display: inline;
          padding: 0;
          margin-left: 0.15rem;
          font-weight: 400;
        }
        .card.inline .header-row {
          display: flex;
          align-items: center;
          gap: 0.4rem;
          flex-wrap: wrap;
        }
      </style>
      <div class="card" id="card">
        <div class="header" id="header">
          <span class="icon" id="icon"></span>
          <span class="title" id="title"></span>
        </div>
        <div class="message" id="message"></div>
        <div class="sections" id="sections" style="display:none"></div>
      </div>
    `}setProps(v){try{let i=v.status||"info",d=I0[i]||I0.info,x=v.sections,$=x&&x.length>0,g=!!v.message,r=this.shadowRoot.getElementById("card"),k=this.shadowRoot.getElementById("header");if(!$&&g){r.classList.add("inline"),k.innerHTML=`
          <div class="header-row">
            <span class="icon" id="icon"></span>
            <span class="title" id="title"></span>
            <span class="message" id="message"></span>
          </div>`;let B=k.querySelector("#icon");B.innerHTML=S[i]||S.info,B.style.color=d;let J=k.querySelector("#title");J.textContent=v.title||"",J.style.color=d,k.querySelector("#message").textContent=v.message,this.shadowRoot.getElementById("message").style.display="none",this.shadowRoot.getElementById("sections").style.display="none",r.style.borderLeft=`3px solid ${d.includes("var(")?d:d}`;return}r.classList.remove("inline");let T=this.shadowRoot.getElementById("icon");T.innerHTML=S[i]||S.info,T.style.color=d;let c=this.shadowRoot.getElementById("title");c.textContent=v.title||"",c.style.color=d;let L=this.shadowRoot.getElementById("message");if(g)L.textContent=v.message,L.style.display="";else L.style.display="none";r.style.borderLeft=`3px solid ${d.includes("var(")?d:d}`;let o=this.shadowRoot.getElementById("sections");if($)o.style.display="",o.innerHTML=x.map((B)=>`
            <div class="section">
              <div class="section-label">${this.esc(B.label||"")}</div>
              <div class="section-content ${B.style==="code"?"code":""}">${this.esc(B.content||"")}</div>
            </div>`).join("");else o.style.display="none"}catch{}}esc(v){return v.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;")}}var N0=15;class T0 extends HTMLElement{allRows=[];filteredRows=[];headers=[];highlight=new Set;searchTerm="";tabs=[];activeTab=0;connectedCallback(){this.attachShadow({mode:"open"}),this.shadowRoot.innerHTML=`
      <style>
        ${R}
        ${z}
        :host {
          display: block;
          animation: fadeSlideUp 0.25s ease-out;
        }
        .card {
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          overflow: hidden;
        }
        .toolbar {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          padding: 0.5rem 0.75rem;
          border-bottom: 1px solid var(--haira-border);
        }
        .toolbar-title {
          font-size: 0.78rem;
          font-weight: 600;
          color: var(--haira-text);
          white-space: nowrap;
        }
        .row-count {
          font-size: 0.68rem;
          color: var(--haira-muted);
          background: var(--haira-bg);
          padding: 0.15rem 0.45rem;
          border-radius: 9px;
          white-space: nowrap;
          flex-shrink: 0;
        }
        .search-wrap {
          margin-left: auto;
          position: relative;
          flex-shrink: 0;
        }
        .search-wrap svg {
          position: absolute;
          left: 0.45rem;
          top: 50%;
          transform: translateY(-50%);
          color: var(--haira-muted);
          pointer-events: none;
        }
        .search {
          background: var(--haira-bg);
          border: 1px solid var(--haira-border);
          color: var(--haira-text);
          font-size: 0.72rem;
          font-family: var(--haira-font);
          padding: 0.28rem 0.5rem 0.28rem 1.6rem;
          border-radius: 6px;
          width: 160px;
          outline: none;
          transition: border-color 0.15s;
        }
        .search:focus { border-color: var(--haira-accent); }
        .search::placeholder { color: var(--haira-muted); }
        .tab-bar {
          display: flex;
          gap: 0;
          border-bottom: 1px solid var(--haira-border);
          background: var(--haira-bg);
          overflow-x: auto;
          scrollbar-width: none;
        }
        .tab-bar::-webkit-scrollbar { display: none; }
        .tab {
          padding: 0.4rem 0.75rem;
          font-size: 0.72rem;
          font-family: var(--haira-font);
          color: var(--haira-muted);
          background: none;
          border: none;
          border-bottom: 2px solid transparent;
          cursor: pointer;
          white-space: nowrap;
          transition: color 0.15s, border-color 0.15s;
          flex-shrink: 0;
        }
        .tab:hover { color: var(--haira-text); }
        .tab.active {
          color: var(--haira-accent);
          border-bottom-color: var(--haira-accent);
          font-weight: 600;
        }
        .table-scroll {
          overflow: auto;
          ${A}
        }
        .table-scroll.capped {
          max-height: 420px;
        }
        table {
          width: 100%;
          border-collapse: collapse;
          font-size: 0.74rem;
        }
        th {
          text-align: left;
          padding: 0.35rem 0.65rem;
          font-weight: 600;
          font-size: 0.68rem;
          color: var(--haira-muted);
          text-transform: uppercase;
          letter-spacing: 0.04em;
          background: var(--haira-bg);
          border-bottom: 1px solid var(--haira-border);
          white-space: nowrap;
          position: sticky;
          top: 0;
          z-index: 1;
        }
        td {
          padding: 0.28rem 0.65rem;
          color: var(--haira-text-dim);
          border-bottom: 1px solid var(--haira-border);
          line-height: 1.35;
          max-width: 320px;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
        td:hover { white-space: normal; word-break: break-all; }
        tr:last-child td { border-bottom: none; }
        tr:hover td { background: var(--haira-bg-card-hover); }
        tr.highlight td { background: rgba(232, 163, 23, 0.06); }
        .no-results {
          padding: 1.5rem;
          text-align: center;
          color: var(--haira-muted);
          font-size: 0.75rem;
        }
        .footer {
          padding: 0.3rem 0.75rem;
          border-top: 1px solid var(--haira-border);
          font-size: 0.68rem;
          color: var(--haira-muted);
          text-align: right;
        }
        @media (max-width: 640px) {
          .toolbar { flex-wrap: wrap; }
          .search-wrap { margin-left: 0; width: 100%; }
          .search { width: 100%; }
          td { max-width: 200px; }
        }
      </style>
      <div class="card">
        <div class="toolbar" id="toolbar" style="display:none">
          <span class="toolbar-title" id="title"></span>
          <span class="row-count" id="row-count"></span>
          <div class="search-wrap" id="search-wrap" style="display:none">
            <svg width="13" height="13" viewBox="0 0 16 16" fill="none"><circle cx="6.5" cy="6.5" r="5" stroke="currentColor" stroke-width="1.5"/><path d="M10.5 10.5L14.5 14.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
            <input class="search" id="search" type="text" placeholder="Filter rows..." />
          </div>
        </div>
        <div class="tab-bar" id="tab-bar" style="display:none"></div>
        <div class="table-scroll" id="scroll">
          <table>
            <thead id="thead"></thead>
            <tbody id="tbody"></tbody>
          </table>
        </div>
        <div class="footer" id="footer" style="display:none"></div>
      </div>
    `,this.shadowRoot.getElementById("search").addEventListener("input",(v)=>{this.searchTerm=v.target.value.toLowerCase(),this.applyFilter()})}setProps(v){try{let{title:i,tabs:d}=v,x=this.shadowRoot.getElementById("toolbar"),$=this.shadowRoot.getElementById("title");if(d&&d.length>0)this.tabs=d,this.activeTab=0,x.style.display="",$.textContent=i||"Table",this.renderTabBar(),this.loadTab(0);else this.tabs=[],this.loadSingleTable(v)}catch{}}loadSingleTable(v){let i=v.title;this.headers=v.headers||[],this.allRows=v.rows||[],this.highlight=new Set(v.highlight||[]);let d=this.shadowRoot.getElementById("toolbar"),x=this.shadowRoot.getElementById("title"),$=this.shadowRoot.getElementById("row-count"),g=this.shadowRoot.getElementById("search-wrap"),r=!!i,k=this.allRows.length>=N0;if(r||k)d.style.display="",x.textContent=i||"Table",$.textContent=`${this.allRows.length} rows`;if(k)g.style.display="";let T=this.shadowRoot.getElementById("scroll");if(k)T.classList.add("capped");else T.classList.remove("capped");let c=this.shadowRoot.getElementById("thead");c.innerHTML=`<tr>${this.headers.map((L)=>`<th>${this.esc(L)}</th>`).join("")}</tr>`,this.searchTerm="",this.shadowRoot.getElementById("search").value="",this.applyFilter()}renderTabBar(){let v=this.shadowRoot.getElementById("tab-bar");v.style.display="flex",v.innerHTML="";for(let i=0;i<this.tabs.length;i++){let d=this.tabs[i],x=document.createElement("button");x.className=`tab${i===this.activeTab?" active":""}`,x.textContent=`${d.name} (${d.rows.length})`,x.addEventListener("click",()=>this.loadTab(i)),v.appendChild(x)}}loadTab(v){this.activeTab=v;let i=this.tabs[v];this.headers=i.headers||[],this.allRows=i.rows||[],this.highlight=new Set(i.highlight||[]),this.shadowRoot.getElementById("tab-bar").querySelectorAll(".tab").forEach((c,L)=>{c.classList.toggle("active",L===v)});let $=this.shadowRoot.getElementById("row-count"),g=this.shadowRoot.getElementById("search-wrap");$.textContent=`${this.allRows.length} rows`;let r=this.allRows.length>=N0;g.style.display=r?"":"none";let k=this.shadowRoot.getElementById("scroll");if(r)k.classList.add("capped");else k.classList.remove("capped");let T=this.shadowRoot.getElementById("thead");T.innerHTML=`<tr>${this.headers.map((c)=>`<th>${this.esc(c)}</th>`).join("")}</tr>`,this.searchTerm="",this.shadowRoot.getElementById("search").value="",this.applyFilter()}applyFilter(){if(this.searchTerm)this.filteredRows=this.allRows.map((v,i)=>({row:v,idx:i})).filter(({row:v})=>v.some((i)=>String(i).toLowerCase().includes(this.searchTerm)));else this.filteredRows=this.allRows.map((v,i)=>({row:v,idx:i}));this.renderRows()}renderRows(){let v=this.shadowRoot.getElementById("tbody"),i=this.shadowRoot.getElementById("footer"),d=this.shadowRoot.getElementById("row-count");if(this.filteredRows.length===0&&this.searchTerm)v.innerHTML=`<tr><td colspan="${this.headers.length||1}" class="no-results">No matching rows</td></tr>`;else v.innerHTML=this.filteredRows.map(({row:x,idx:$})=>`<tr class="${this.highlight.has($)?"highlight":""}">${x.map((g)=>`<td title="${this.esc(String(g))}">${this.esc(String(g))}</td>`).join("")}</tr>`).join("");if(this.searchTerm)d.textContent=`${this.filteredRows.length} / ${this.allRows.length} rows`,i.style.display="",i.textContent=`Showing ${this.filteredRows.length} of ${this.allRows.length} rows`;else d.textContent=`${this.allRows.length} rows`,i.style.display="none"}esc(v){return v.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;").replace(/"/g,"&quot;")}}var P0='<svg width="13" height="13" viewBox="0 0 16 16" fill="none"><rect x="5" y="5" width="9" height="9" rx="1.5" stroke="currentColor" stroke-width="1.5"/><path d="M11 5V3.5A1.5 1.5 0 0 0 9.5 2H3.5A1.5 1.5 0 0 0 2 3.5v6A1.5 1.5 0 0 0 3.5 11H5" stroke="currentColor" stroke-width="1.5"/></svg>',x1='<svg width="13" height="13" viewBox="0 0 16 16" fill="none"><path d="M4 8.5L6.5 11L12 5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>';class c0 extends HTMLElement{codeText="";tabs=[];activeTab=0;connectedCallback(){this.attachShadow({mode:"open"}),this.shadowRoot.innerHTML=`
      <style>
        ${R}
        ${z}
        :host {
          display: block;
          animation: fadeSlideUp 0.25s ease-out;
        }
        .card {
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          overflow: hidden;
        }
        .header {
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: 0.45rem 0.75rem;
          border-bottom: 1px solid var(--haira-border);
          background: var(--haira-bg);
        }
        .title {
          font-size: 0.78rem;
          font-weight: 600;
          color: var(--haira-text);
        }
        .lang {
          font-size: 0.68rem;
          color: var(--haira-muted);
          font-family: var(--haira-mono);
        }
        .actions {
          display: flex;
          align-items: center;
          gap: 0.5rem;
        }
        .copy-btn {
          background: none;
          border: none;
          color: var(--haira-muted);
          cursor: pointer;
          display: flex;
          align-items: center;
          gap: 0.3rem;
          font-size: 0.7rem;
          font-family: var(--haira-font);
          padding: 0.2rem 0.4rem;
          border-radius: 4px;
          transition: all 0.15s;
        }
        .copy-btn:hover { color: var(--haira-accent); background: var(--haira-accent-dim); }
        .copy-btn.copied { color: var(--haira-success); }
        .tab-bar {
          display: flex;
          gap: 0;
          border-bottom: 1px solid var(--haira-border);
          background: var(--haira-bg);
          overflow-x: auto;
          scrollbar-width: none;
        }
        .tab-bar::-webkit-scrollbar { display: none; }
        .tab {
          padding: 0.4rem 0.75rem;
          font-size: 0.72rem;
          font-family: var(--haira-font);
          color: var(--haira-muted);
          background: none;
          border: none;
          border-bottom: 2px solid transparent;
          cursor: pointer;
          white-space: nowrap;
          transition: color 0.15s, border-color 0.15s;
          flex-shrink: 0;
        }
        .tab:hover { color: var(--haira-text); }
        .tab.active {
          color: var(--haira-accent);
          border-bottom-color: var(--haira-accent);
          font-weight: 600;
        }
        .code-scroll {
          max-height: 480px;
          overflow: auto;
          ${A}
        }
        pre {
          margin: 0;
          padding: 0.75rem 1rem;
        }
        code {
          font-family: var(--haira-mono);
          font-size: 0.78rem;
          color: var(--haira-text-dim);
          line-height: 1.6;
          white-space: pre;
        }
      </style>
      <div class="card">
        <div class="header">
          <div style="display:flex;align-items:center;gap:0.5rem">
            <span class="title" id="title"></span>
            <span class="lang" id="lang"></span>
          </div>
          <div class="actions">
            <button class="copy-btn" id="copy-btn">${P0} Copy</button>
          </div>
        </div>
        <div class="tab-bar" id="tab-bar" style="display:none"></div>
        <div class="code-scroll" id="code-scroll">
          <pre><code id="code"></code></pre>
        </div>
      </div>
    `,this.shadowRoot.getElementById("copy-btn").addEventListener("click",()=>{this.copyCode()})}setProps(v){try{let i=this.shadowRoot.getElementById("title");i.textContent=v.title||"";let d=v.tabs;if(d&&d.length>0)this.tabs=d,this.activeTab=0,this.renderTabBar(),this.loadTab(0);else{this.tabs=[];let x=this.shadowRoot.getElementById("lang");x.textContent=v.language||"";let $=this.shadowRoot.getElementById("code");this.codeText=v.code||"",$.textContent=this.codeText}}catch{}}renderTabBar(){let v=this.shadowRoot.getElementById("tab-bar");v.style.display="flex",v.innerHTML="";for(let i=0;i<this.tabs.length;i++){let d=this.tabs[i],x=document.createElement("button");x.className=`tab${i===this.activeTab?" active":""}`,x.textContent=d.name,x.addEventListener("click",()=>this.loadTab(i)),v.appendChild(x)}}loadTab(v){this.activeTab=v;let i=this.tabs[v];this.shadowRoot.getElementById("tab-bar").querySelectorAll(".tab").forEach((g,r)=>{g.classList.toggle("active",r===v)});let x=this.shadowRoot.getElementById("lang");x.textContent=i.language||"",this.codeText=i.code||"";let $=this.shadowRoot.getElementById("code");$.textContent=this.codeText,this.shadowRoot.getElementById("code-scroll").scrollTop=0}copyCode(){navigator.clipboard.writeText(this.codeText).then(()=>{let v=this.shadowRoot.getElementById("copy-btn");v.innerHTML=`${x1} Copied`,v.classList.add("copied"),setTimeout(()=>{v.innerHTML=`${P0} Copy`,v.classList.remove("copied")},2000)})}}class R0 extends HTMLElement{connectedCallback(){this.attachShadow({mode:"open"}),this.shadowRoot.innerHTML=`
      <style>
        ${R}
        ${z}
        :host {
          display: block;
          animation: fadeSlideUp 0.25s ease-out;
        }
        .card {
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          overflow: hidden;
        }
        .title-bar {
          padding: 0.6rem 1rem;
          font-size: 0.8rem;
          font-weight: 600;
          color: var(--haira-text);
          border-bottom: 1px solid var(--haira-border);
          display: none;
        }
        .diff-grid {
          display: grid;
          grid-template-columns: 1fr 1fr;
        }
        .pane {
          overflow-x: auto;
          ${A}
        }
        .pane + .pane {
          border-left: 1px solid var(--haira-border);
        }
        .pane-header {
          padding: 0.4rem 0.75rem;
          font-size: 0.72rem;
          font-weight: 600;
          color: var(--haira-muted);
          text-transform: uppercase;
          letter-spacing: 0.04em;
          background: var(--haira-bg);
          border-bottom: 1px solid var(--haira-border);
        }
        .pane-header.before { color: var(--haira-error); }
        .pane-header.after { color: var(--haira-success); }
        pre {
          margin: 0;
          padding: 0.6rem 0.75rem;
          font-family: var(--haira-mono);
          font-size: 0.75rem;
          color: var(--haira-text-dim);
          line-height: 1.6;
          white-space: pre;
          min-height: 3rem;
        }
        .pane:first-child pre {
          background: rgba(239, 68, 68, 0.03);
        }
        .pane:last-child pre {
          background: rgba(34, 197, 94, 0.03);
        }
      </style>
      <div class="card">
        <div class="title-bar" id="title"></div>
        <div class="diff-grid">
          <div class="pane">
            <div class="pane-header before" id="before-label">Before</div>
            <pre id="before"></pre>
          </div>
          <div class="pane">
            <div class="pane-header after" id="after-label">After</div>
            <pre id="after"></pre>
          </div>
        </div>
      </div>
    `}setProps(v){try{let i=this.shadowRoot.getElementById("title");if(v.title)i.textContent=v.title,i.style.display="";let d=this.shadowRoot.getElementById("before-label");d.textContent=v.before_label||"Before";let x=this.shadowRoot.getElementById("after-label");x.textContent=v.after_label||"After",this.shadowRoot.getElementById("before").textContent=v.before||"",this.shadowRoot.getElementById("after").textContent=v.after||""}catch{}}}var $1={success:"var(--haira-success)",error:"var(--haira-error)",warning:"var(--haira-warn)",info:"var(--haira-info)",code:"inherit"};class o0 extends HTMLElement{connectedCallback(){this.attachShadow({mode:"open"}),this.shadowRoot.innerHTML=`
      <style>
        ${R}
        ${z}
        :host {
          display: block;
          animation: fadeSlideUp 0.25s ease-out;
        }
        .card {
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          overflow: hidden;
        }
        .title-bar {
          padding: 0.6rem 1rem;
          font-size: 0.8rem;
          font-weight: 600;
          color: var(--haira-text);
          border-bottom: 1px solid var(--haira-border);
          display: none;
        }
        .items {
          padding: 0.5rem 0;
        }
        .item {
          display: flex;
          align-items: baseline;
          padding: 0.3rem 1rem;
          gap: 0.75rem;
        }
        .key {
          font-size: 0.75rem;
          font-weight: 600;
          color: var(--haira-muted);
          min-width: 100px;
          flex-shrink: 0;
        }
        .value {
          font-size: 0.8rem;
          color: var(--haira-text-dim);
          word-break: break-word;
        }
        .value.code {
          font-family: var(--haira-mono);
          font-size: 0.75rem;
          background: var(--haira-bg);
          padding: 0.15rem 0.4rem;
          border-radius: 4px;
        }
      </style>
      <div class="card">
        <div class="title-bar" id="title"></div>
        <div class="items" id="items"></div>
      </div>
    `}setProps(v){try{let i=this.shadowRoot.getElementById("title");if(v.title)i.textContent=v.title,i.style.display="";let d=v.items||[],x=this.shadowRoot.getElementById("items");x.innerHTML=d.map(($)=>{let g=$1[$.style]||"",r=g&&g!=="inherit"?`color:${g}`:"",k=$.style==="code";return`<div class="item">
            <span class="key">${this.esc($.key||"")}</span>
            <span class="value ${k?"code":""}" style="${r}">${this.esc($.value||"")}</span>
          </div>`}).join("")}catch{}}esc(v){return v.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;")}}var _0={done:{icon:'<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="7" fill="currentColor" opacity="0.15"/><path d="M5 8.5L7 10.5L11 5.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>',color:"var(--haira-success)"},active:{icon:'<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-dasharray="28 10" style="animation:spin 0.7s linear infinite;transform-origin:center"/></svg>',color:"var(--haira-accent)"},pending:{icon:'<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="6.5" stroke="currentColor" stroke-width="1.5" stroke-dasharray="3 2"/></svg>',color:"var(--haira-muted)"},failed:{icon:'<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="7" fill="currentColor" opacity="0.15"/><path d="M5.5 5.5L10.5 10.5M10.5 5.5L5.5 10.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>',color:"var(--haira-error)"}};class z0 extends HTMLElement{connectedCallback(){this.attachShadow({mode:"open"}),this.shadowRoot.innerHTML=`
      <style>
        ${R}
        ${z}
        :host {
          display: block;
          animation: fadeSlideUp 0.25s ease-out;
        }
        .card {
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          overflow: hidden;
        }
        .title-bar {
          padding: 0.6rem 1rem;
          font-size: 0.8rem;
          font-weight: 600;
          color: var(--haira-text);
          border-bottom: 1px solid var(--haira-border);
          display: none;
        }
        .steps {
          padding: 0.6rem 1rem;
        }
        .step {
          display: flex;
          align-items: flex-start;
          gap: 0.6rem;
          position: relative;
          padding-bottom: 0.75rem;
        }
        .step:last-child { padding-bottom: 0; }
        .step::before {
          content: "";
          position: absolute;
          left: 6.5px;
          top: 18px;
          bottom: 0;
          width: 1px;
          background: var(--haira-border);
        }
        .step:last-child::before { display: none; }
        .step-icon { display: flex; flex-shrink: 0; margin-top: 1px; }
        .step-content { flex: 1; min-width: 0; }
        .step-name {
          font-size: 0.8rem;
          font-weight: 500;
          line-height: 1.3;
        }
        .step-detail {
          font-size: 0.72rem;
          color: var(--haira-muted);
          margin-top: 0.15rem;
        }
        @keyframes spin {
          to { transform: rotate(360deg); }
        }
      </style>
      <div class="card">
        <div class="title-bar" id="title"></div>
        <div class="steps" id="steps"></div>
      </div>
    `}setProps(v){try{let i=this.shadowRoot.getElementById("title");if(v.title)i.textContent=v.title,i.style.display="";let d=v.steps||[],x=this.shadowRoot.getElementById("steps");x.innerHTML=d.map(($)=>{let g=$.status||"pending",r=_0[g]||_0.pending;return`<div class="step">
            <span class="step-icon" style="color:${r.color}">${r.icon}</span>
            <div class="step-content">
              <div class="step-name" style="color:${r.color}">${this.esc($.name||"")}</div>
              ${$.detail?`<div class="step-detail">${this.esc($.detail)}</div>`:""}
            </div>
          </div>`}).join("")}catch{}}esc(v){return v.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;")}}class L0 extends HTMLElement{connectedCallback(){this.attachShadow({mode:"open"}),this.shadowRoot.innerHTML=`
      <style>
        ${R}
        ${z}
        :host {
          display: block;
          animation: fadeSlideUp 0.25s ease-out;
        }
        .card {
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          overflow: hidden;
        }
        .title-bar {
          padding: 0.6rem 1rem;
          font-size: 0.8rem;
          font-weight: 600;
          color: var(--haira-text);
          border-bottom: 1px solid var(--haira-border);
          display: none;
        }
        .fields {
          padding: 0.75rem 1rem;
          display: flex;
          flex-direction: column;
          gap: 0.6rem;
        }
        .field-label {
          font-size: 0.75rem;
          font-weight: 600;
          color: var(--haira-text-dim);
          margin-bottom: 0.2rem;
        }
        .field-label .required {
          color: var(--haira-error);
          margin-left: 0.2rem;
        }
        input, select, textarea {
          width: 100%;
          background: var(--haira-bg);
          border: 1px solid var(--haira-border);
          color: var(--haira-text);
          padding: 0.45rem 0.65rem;
          border-radius: var(--haira-radius-sm);
          font-size: 0.8rem;
          font-family: var(--haira-font);
          outline: none;
          transition: border-color 0.15s;
        }
        input:focus, select:focus, textarea:focus {
          border-color: var(--haira-border-focus);
        }
        textarea { min-height: 60px; resize: vertical; }
        select {
          cursor: pointer;
          appearance: none;
          background-image: url("data:image/svg+xml,%3Csvg width='10' height='6' viewBox='0 0 10 6' fill='none' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M1 1l4 4 4-4' stroke='%2371717a' stroke-width='1.5' stroke-linecap='round'/%3E%3C/svg%3E");
          background-repeat: no-repeat;
          background-position: right 0.6rem center;
          padding-right: 2rem;
        }
        .submit-area {
          padding: 0.5rem 1rem 0.75rem;
          border-top: 1px solid var(--haira-border);
        }
        .submit-btn {
          background: var(--haira-accent);
          color: #1a0e04;
          border: none;
          padding: 0.5rem 1.2rem;
          border-radius: var(--haira-radius-sm);
          font-size: 0.8rem;
          font-weight: 600;
          font-family: var(--haira-font);
          cursor: pointer;
          transition: all 0.15s;
        }
        .submit-btn:hover {
          background: var(--haira-accent-light);
          box-shadow: 0 2px 12px rgba(232, 163, 23, 0.25);
        }
      </style>
      <div class="card">
        <div class="title-bar" id="title"></div>
        <div class="fields" id="fields"></div>
        <div class="submit-area">
          <button class="submit-btn" id="submit-btn">Submit</button>
        </div>
      </div>
    `}setProps(v){try{let i=this.shadowRoot.getElementById("title");if(v.title)i.textContent=v.title,i.style.display="";let d=this.shadowRoot.getElementById("submit-btn");d.textContent=v.submit_label||"Submit";let x=v.fields||[],$=this.shadowRoot.getElementById("fields");$.innerHTML=x.map((g)=>{let r=g.name||"",k=g.label||r,T=g.field_type||"text",c=g.value||"",L=g.required,o=g.options||[],B;if(T==="select"&&o.length>0)B=`<select name="${this.esc(r)}">
              ${o.map((J)=>`<option value="${this.esc(J)}" ${J===c?"selected":""}>${this.esc(J)}</option>`).join("")}
            </select>`;else if(T==="textarea")B=`<textarea name="${this.esc(r)}">${this.esc(c)}</textarea>`;else B=`<input type="${this.esc(T)}" name="${this.esc(r)}" value="${this.esc(c)}" ${L?"required":""} />`;return`<div class="field-group">
            <div class="field-label">${this.esc(k)}${L?'<span class="required">*</span>':""}</div>
            ${B}
          </div>`}).join(""),d.onclick=()=>{let g={};$.querySelectorAll("input, select, textarea").forEach((r)=>{let k=r;g[k.name]=k.value}),this.dispatchEvent(new CustomEvent("haira-form-submit",{detail:{action:v.submit_action||"",data:g},bubbles:!0,composed:!0}))}}catch{}}esc(v){return v.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;").replace(/"/g,"&quot;")}}class B0 extends HTMLElement{answered=!1;connectedCallback(){this.attachShadow({mode:"open"}),this.shadowRoot.innerHTML=`
      <style>
        ${R}
        ${z}
        :host {
          display: block;
          animation: fadeSlideUp 0.25s ease-out;
        }
        .card {
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-left: 3px solid var(--haira-info);
          border-radius: var(--haira-radius);
          padding: 0.55rem 0.75rem;
        }
        .title {
          font-size: 0.78rem;
          font-weight: 600;
          color: var(--haira-text);
          margin-bottom: 0.15rem;
        }
        .message {
          font-size: 0.73rem;
          color: var(--haira-text-dim);
          line-height: 1.4;
          margin-bottom: 0.5rem;
        }
        .actions {
          display: flex;
          gap: 0.4rem;
        }
        button {
          font-family: var(--haira-font);
          font-size: 0.73rem;
          font-weight: 600;
          padding: 0.32rem 0.85rem;
          border-radius: 6px;
          cursor: pointer;
          transition: all 0.15s;
        }
        .confirm-btn {
          background: var(--haira-accent);
          color: #1a0e04;
          border: none;
        }
        .confirm-btn:hover { background: var(--haira-accent-light); }
        .deny-btn {
          background: transparent;
          color: var(--haira-text-dim);
          border: 1px solid var(--haira-border);
        }
        .deny-btn:hover {
          border-color: var(--haira-text-dim);
          color: var(--haira-text);
        }
        button:disabled {
          opacity: 0.4;
          cursor: default;
          pointer-events: none;
        }
        button.selected {
          opacity: 1;
        }
        .selected-indicator {
          display: none;
          font-size: 0.68rem;
          color: var(--haira-muted);
          margin-top: 0.35rem;
        }
        .selected-indicator.visible { display: block; }
      </style>
      <div class="card" id="card">
        <div class="title" id="title"></div>
        <div class="message" id="message"></div>
        <div class="actions" id="actions">
          <button class="confirm-btn" id="confirm-btn"></button>
          <button class="deny-btn" id="deny-btn"></button>
        </div>
        <div class="selected-indicator" id="indicator"></div>
      </div>
    `}setProps(v){try{let i=this.shadowRoot.getElementById("title");i.textContent=v.title||"Confirm";let d=this.shadowRoot.getElementById("message");if(v.message)d.textContent=v.message,d.style.display="";else d.style.display="none";let x=v.confirm_label||"Confirm",$=v.deny_label||"Cancel",g=this.shadowRoot.getElementById("confirm-btn"),r=this.shadowRoot.getElementById("deny-btn");if(g.textContent=x,r.textContent=$,v._restored)this.answered=!0,g.disabled=!0,r.disabled=!0;else g.onclick=()=>this.select(x,g,r),r.onclick=()=>this.select($,r,g)}catch{}}select(v,i,d){if(this.answered)return;this.answered=!0,i.classList.add("selected"),i.disabled=!0,d.disabled=!0;let x=this.shadowRoot.getElementById("indicator");x.textContent=`Selected: ${v}`,x.classList.add("visible");let $=this.shadowRoot.getElementById("title")?.textContent||"",g=`[User clicked "${v}" on confirmation: ${$}]`;this.dispatchEvent(new CustomEvent("haira-chat-input",{detail:{text:g},bubbles:!0,composed:!0}))}}class Q0 extends HTMLElement{answered=!1;connectedCallback(){this.attachShadow({mode:"open"}),this.shadowRoot.innerHTML=`
      <style>
        ${R}
        ${z}
        :host {
          display: block;
          animation: fadeSlideUp 0.25s ease-out;
        }
        .card {
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          padding: 0.55rem 0.75rem;
        }
        .title {
          font-size: 0.78rem;
          font-weight: 600;
          color: var(--haira-text);
          margin-bottom: 0.45rem;
        }

        /* Buttons style (default) */
        .options-buttons {
          display: flex;
          flex-wrap: wrap;
          gap: 0.35rem;
        }
        .opt-btn {
          background: transparent;
          border: 1px solid var(--haira-border);
          color: var(--haira-text-dim);
          font-family: var(--haira-font);
          font-size: 0.73rem;
          padding: 0.3rem 0.7rem;
          border-radius: 16px;
          cursor: pointer;
          transition: all 0.15s;
        }
        .opt-btn:hover {
          border-color: var(--haira-accent);
          color: var(--haira-accent);
          background: var(--haira-accent-dim);
        }
        .opt-btn:disabled {
          opacity: 0.35;
          cursor: default;
          pointer-events: none;
        }
        .opt-btn.selected {
          opacity: 1;
          background: var(--haira-accent);
          color: #1a0e04;
          border-color: var(--haira-accent);
        }

        /* List style */
        .options-list {
          display: flex;
          flex-direction: column;
          gap: 0.15rem;
        }
        .opt-row {
          display: flex;
          align-items: center;
          gap: 0.45rem;
          padding: 0.35rem 0.5rem;
          border-radius: 6px;
          cursor: pointer;
          transition: background 0.15s;
          font-size: 0.75rem;
          color: var(--haira-text-dim);
        }
        .opt-row:hover { background: var(--haira-bg-card-hover); }
        .opt-radio {
          width: 14px;
          height: 14px;
          border-radius: 50%;
          border: 2px solid var(--haira-border);
          flex-shrink: 0;
          transition: all 0.15s;
          display: flex;
          align-items: center;
          justify-content: center;
        }
        .opt-row:hover .opt-radio { border-color: var(--haira-accent); }
        .opt-row.selected .opt-radio {
          border-color: var(--haira-accent);
          background: var(--haira-accent);
        }
        .opt-row.selected .opt-radio::after {
          content: "";
          width: 5px;
          height: 5px;
          border-radius: 50%;
          background: #1a0e04;
        }
        .opt-row.disabled {
          opacity: 0.35;
          cursor: default;
          pointer-events: none;
        }
        .opt-row.selected.disabled {
          opacity: 1;
        }
      </style>
      <div class="card">
        <div class="title" id="title"></div>
        <div id="options"></div>
      </div>
    `}setProps(v){try{let i=this.shadowRoot.getElementById("title");i.textContent=v.title||"Choose an option";let d=v.options||[],x=v.style||"buttons",$=this.shadowRoot.getElementById("options"),g=!!v._restored;if(g)this.answered=!0;if(x==="list"){if($.className="options-list",$.innerHTML=d.map((r)=>`<div class="opt-row${g?" disabled":""}" data-value="${this.esc(r)}">
                <span class="opt-radio"></span>
                <span>${this.esc(r)}</span>
              </div>`).join(""),!g)$.querySelectorAll(".opt-row").forEach((r)=>{r.addEventListener("click",()=>{this.selectOption(r.dataset.value||"",$,"list")})})}else if($.className="options-buttons",$.innerHTML=d.map((r)=>`<button class="opt-btn" data-value="${this.esc(r)}"${g?" disabled":""}>${this.esc(r)}</button>`).join(""),!g)$.querySelectorAll(".opt-btn").forEach((r)=>{r.addEventListener("click",()=>{this.selectOption(r.dataset.value||"",$,"buttons")})})}catch{}}selectOption(v,i,d){if(this.answered)return;if(this.answered=!0,d==="list")i.querySelectorAll(".opt-row").forEach((g)=>{let r=g;if(r.dataset.value===v)r.classList.add("selected","disabled");else r.classList.add("disabled")});else i.querySelectorAll(".opt-btn").forEach((g)=>{let r=g;if(r.dataset.value===v)r.classList.add("selected");r.disabled=!0});let x=this.shadowRoot.getElementById("title")?.textContent||"",$=`[User selected "${v}" from choices: ${x}]`;this.dispatchEvent(new CustomEvent("haira-chat-input",{detail:{text:$},bubbles:!0,composed:!0}))}esc(v){return v.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;").replace(/"/g,"&quot;")}}var r1={table:"haira-ui-table","status-card":"haira-ui-status-card","code-block":"haira-ui-code-block",diff:"haira-ui-diff","key-value":"haira-ui-key-value",progress:"haira-ui-progress-view",form:"haira-ui-form-view",confirm:"haira-ui-confirm",choices:"haira-ui-choices"},k1=3;class Y0 extends HTMLElement{connectedCallback(){this.attachShadow({mode:"open"}),this.shadowRoot.innerHTML=`
      <style>
        ${R}
        :host {
          display: block;
          margin-left: 2.25rem;
          max-width: 560px;
        }
        @media (max-width: 640px) {
          :host {
            margin-left: 0;
            max-width: 100%;
          }
        }
        .group {
          display: flex;
          flex-direction: column;
          gap: 0.5rem;
        }
        .fallback {
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          padding: 0.75rem 1rem;
          font-family: var(--haira-mono);
          font-size: 0.75rem;
          color: var(--haira-text-dim);
          white-space: pre-wrap;
          overflow-x: auto;
        }
      </style>
      <div id="container"></div>
    `}render(v){let i=this.shadowRoot.getElementById("container");i.innerHTML="";try{let d=this.renderNode(v.component,v.props,0);if(d)i.appendChild(d)}catch{let d=document.createElement("div");d.className="fallback",d.textContent=JSON.stringify(v.props,null,2),i.appendChild(d)}}renderNode(v,i,d){if(d>k1)return null;if(v==="group"){let k=document.createElement("div");k.className="group";let T=i.children||[];for(let c of T){let{component:L,props:o}=c;if(L&&o){let B=this.renderNode(L,o,d+1);if(B)k.appendChild(B)}}return k}let x=r1[v];if(!x){let k=document.createElement("div");return k.className="fallback",k.textContent=JSON.stringify(i,null,2),k}let $=document.createElement(x),r=this.hasAttribute("data-restored")?{...i,_restored:!0}:i;return requestAnimationFrame(()=>{$.setProps(r)}),$}}customElements.define("haira-field",s);customElements.define("haira-result",e);customElements.define("haira-step",v0);customElements.define("haira-pipeline",i0);customElements.define("haira-message",d0);customElements.define("haira-tool-card",r0);customElements.define("haira-ui-status-card",k0);customElements.define("haira-ui-table",T0);customElements.define("haira-ui-code-block",c0);customElements.define("haira-ui-diff",R0);customElements.define("haira-ui-key-value",o0);customElements.define("haira-ui-progress-view",z0);customElements.define("haira-ui-form-view",L0);customElements.define("haira-ui-confirm",B0);customElements.define("haira-ui-choices",Q0);customElements.define("haira-ui-renderer",Y0);customElements.define("haira-form",g0);customElements.define("haira-index",x0);customElements.define("haira-chat",$0);customElements.define("haira-app",p);
