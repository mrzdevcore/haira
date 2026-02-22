var ae=`
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
`;var $=`
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

`,R=`
  ::-webkit-scrollbar { width: 5px; height: 5px; }
  ::-webkit-scrollbar-track { background: transparent; }
  ::-webkit-scrollbar-thumb { background: var(--haira-muted); border-radius: 3px; }
  ::-webkit-scrollbar-thumb:hover { background: var(--haira-accent); }
  scrollbar-width: thin;
  scrollbar-color: var(--haira-muted) transparent;
`,T=`
  background: var(--haira-bg-card);
  border: 1px solid var(--haira-border);
  border-radius: var(--haira-radius);
  overflow: hidden;
`,k=`
  :host {
    display: block;
    animation: fadeSlideUp 0.25s ease-out;
  }
`;function D(e){switch(e){case"POST":return"#22c55e";case"GET":return"#3b82f6";case"PUT":return"#f59e0b";case"DELETE":return"#ef4444";default:return"#71717a"}}function Y(e){switch(e){case"form":return"#3b82f6";case"chat":return"#22c55e";case"stream":return"#e8a317";default:return"#71717a"}}function Q(e){let t=e.match(/^#?([0-9a-f]{6})$/i);if(!t)return null;return{r:parseInt(t[1].substring(0,2),16),g:parseInt(t[1].substring(2,4),16),b:parseInt(t[1].substring(4,6),16)}}function ie(e,t){let r=Q(e);if(!r)return e;let i=Math.min(255,r.r+Math.round((255-r.r)*t)),o=Math.min(255,r.g+Math.round((255-r.g)*t)),a=Math.min(255,r.b+Math.round((255-r.b)*t));return`#${i.toString(16).padStart(2,"0")}${o.toString(16).padStart(2,"0")}${a.toString(16).padStart(2,"0")}`}function p(e){return e.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;")}function oe(e){return p(e).replace(/"/g,"&quot;")}function ne(e){if(e<1024)return`${e} B`;if(e<1048576)return`${(e/1024).toFixed(1)} KB`;return`${(e/1048576).toFixed(1)} MB`}function se(e){return e.replace(/_/g," ").replace(/\b\w/g,(t)=>t.toUpperCase())}class w extends HTMLElement{root;props={};connectedCallback(){this.root=this.attachShadow({mode:"open"}),this.root.innerHTML=`<style>${$}
${this.styles()}</style>${this.render()}`,this.onMount()}setProps(e){this.props=e,this.onUpdate()}styles(){return""}onMount(){}onUpdate(){}$(e){return this.root.getElementById(e)}$q(e){return this.root.querySelector(e)}$qa(e){return this.root.querySelectorAll(e)}esc(e){return p(e)}escAttr(e){return oe(e)}emit(e,t){this.dispatchEvent(new CustomEvent(e,{detail:t,bubbles:!0,composed:!0}))}}var h={pending:'<svg width="16" height="16" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="6.5" stroke="currentColor" stroke-width="1.5" stroke-dasharray="3 2"/></svg>',spinner:'<svg width="16" height="16" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-dasharray="28 10" style="animation:spin 0.7s linear infinite;transform-origin:center"/></svg>',check:'<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><path d="M4.5 8.5L7 11L11.5 5.5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>',x:'<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><path d="M5 5L11 11M11 5L5 11" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>',retry:'<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><path d="M3 8a5 5 0 0 1 8.5-3.5M13 8a5 5 0 0 1-8.5 3.5" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/><path d="M11 2v3h-3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/><path d="M5 14v-3h3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>',copy:'<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><rect x="5" y="5" width="9" height="9" rx="1.5" stroke="currentColor" stroke-width="1.5"/><path d="M11 5V3.5A1.5 1.5 0 0 0 9.5 2H3.5A1.5 1.5 0 0 0 2 3.5v6A1.5 1.5 0 0 0 3.5 11H5" stroke="currentColor" stroke-width="1.5"/></svg>',copyDone:'<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><path d="M4 8.5L6.5 11L12 5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>',chevron:'<svg width="12" height="12" viewBox="0 0 16 16" fill="none"><path d="M6 4l4 4-4 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>',chevronRight:'<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>',chevronLeft:'<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 18 9 12 15 6"/></svg>',send:'<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>',attach:'<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21.44 11.05l-9.19 9.19a6 6 0 01-8.49-8.49l9.19-9.19a4 4 0 015.66 5.66l-9.2 9.19a2 2 0 01-2.83-2.83l8.49-8.48"/></svg>',file:'<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>',xSmall:'<svg width="12" height="12" viewBox="0 0 16 16" fill="none"><path d="M5 5L11 11M11 5L5 11" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>',activity:'<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>',plus:'<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>',chat:'<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z"/></svg>',trash:'<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/></svg>',sidebar:'<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2"/><line x1="9" y1="3" x2="9" y2="21"/></svg>',search:'<svg width="13" height="13" viewBox="0 0 16 16" fill="none"><circle cx="6.5" cy="6.5" r="5" stroke="currentColor" stroke-width="1.5"/><path d="M10.5 10.5L14.5 14.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>',tool:'<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14.7 6.3a1 1 0 000 1.4l1.6 1.6a1 1 0 001.4 0l3.77-3.77a6 6 0 01-7.94 7.94l-6.91 6.91a2.12 2.12 0 01-3-3l6.91-6.91a6 6 0 017.94-7.94l-3.76 3.76z"/></svg>',statusSuccess:'<svg width="16" height="16" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="7" stroke="currentColor" stroke-width="1.5"/><path d="M5 8.5L7 10.5L11 5.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>',statusError:'<svg width="16" height="16" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="7" stroke="currentColor" stroke-width="1.5"/><path d="M5.5 5.5L10.5 10.5M10.5 5.5L5.5 10.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>',statusWarning:'<svg width="16" height="16" viewBox="0 0 16 16" fill="none"><path d="M8 2L14.5 13H1.5L8 2Z" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"/><path d="M8 6.5V9" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><circle cx="8" cy="11" r="0.75" fill="currentColor"/></svg>',statusInfo:'<svg width="16" height="16" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="7" stroke="currentColor" stroke-width="1.5"/><path d="M8 7V11" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><circle cx="8" cy="5" r="0.75" fill="currentColor"/></svg>',stepDone:'<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="7" fill="currentColor" opacity="0.15"/><path d="M5 8.5L7 10.5L11 5.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>',stepActive:'<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-dasharray="28 10" style="animation:spin 0.7s linear infinite;transform-origin:center"/></svg>',stepPending:'<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="6.5" stroke="currentColor" stroke-width="1.5" stroke-dasharray="3 2"/></svg>',stepFailed:'<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="7" fill="currentColor" opacity="0.15"/><path d="M5.5 5.5L10.5 10.5M10.5 5.5L5.5 10.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>'};var q='<svg width="22" height="22" viewBox="0 0 64 52" fill="none" xmlns="http://www.w3.org/2000/svg"><rect x="17" y="11" width="30" height="20" rx="6" fill="#F0BD4F"/><rect x="21" y="15" width="22" height="9" rx="4" fill="#3D2B1F"/><circle cx="27" cy="19.5" r="3.5" fill="#FDE68A"/><circle cx="27" cy="19.5" r="1.5" fill="#fff"/><circle cx="37" cy="19.5" r="3.5" fill="#FDE68A"/><circle cx="37" cy="19.5" r="1.5" fill="#fff"/><ellipse cx="32" cy="12" rx="25" ry="4" fill="#C4A265"/><ellipse cx="32" cy="11.5" rx="23" ry="3" fill="#D4B87A"/><rect x="18" y="2" width="28" height="10" rx="5" fill="#C4A265"/><rect x="20" y="1" width="24" height="5" rx="3" fill="#D4B87A"/><rect x="18" y="8" width="28" height="3.5" rx="1.5" fill="#5C3A1E"/><rect x="20" y="31" width="24" height="14" rx="4" fill="#E8A317"/><rect x="25" y="34" width="14" height="8" rx="3" fill="#3D2B1F"/><circle cx="32" cy="38" r="3" fill="#FDE68A"/><rect x="10" y="34" width="10" height="4" rx="2" fill="#E8A317"/><rect x="44" y="34" width="10" height="4" rx="2" fill="#E8A317"/></svg>';function _e(e,t){if(t.theme==="light")for(let i of ae.split(`
`)){let o=i.match(/(--[\w-]+):\s*(.+);/);if(o)e.style.setProperty(o[1],o[2].trim())}let r=t.accent;if(r){e.style.setProperty("--haira-accent",r),e.style.setProperty("--haira-accent-light",ie(r,0.25));let i=Q(r);if(i)e.style.setProperty("--haira-accent-dim",`rgba(${i.r}, ${i.g}, ${i.b}, 0.06)`),e.style.setProperty("--haira-border-light",`rgba(${i.r}, ${i.g}, ${i.b}, 0.12)`),e.style.setProperty("--haira-border-focus",`rgba(${i.r}, ${i.g}, ${i.b}, 0.4)`)}}class le extends HTMLElement{meta=null;connectedCallback(){let e=document.getElementById("haira-meta");if(e)try{this.meta=JSON.parse(e.textContent||"{}")}catch{this.meta=null}if(this.meta){let t=this.meta.title||this.meta.name;document.title=t?`${t} — Haira`:"Haira"}this.renderApp()}renderApp(){let e=this.attachShadow({mode:"open"});e.innerHTML=`
      <style>
        ${$}
        :host { display: block; height: 100vh; overflow: hidden; background: var(--haira-bg); }
        .shell { height: 100%; display: flex; flex-direction: column; overflow: hidden; }
        header {
          padding: 0.6rem 1.25rem; border-bottom: 1px solid var(--haira-border);
          display: flex; align-items: center; gap: 0.6rem;
          background: var(--haira-bg); position: sticky; top: 0; z-index: 100;
        }
        .logo { display: flex; align-items: center; gap: 0.4rem; text-decoration: none; }
        .logo-icon { display: flex; align-items: center; }
        .logo-icon img { width: 22px; height: 22px; object-fit: contain; }
        .logo-text { font-weight: 700; font-size: 0.92rem; color: var(--haira-text); letter-spacing: -0.01em; }
        .logo-text .ai { color: var(--haira-accent); }
        .sep { color: var(--haira-muted); font-size: 0.75rem; opacity: 0.5; }
        .title { color: var(--haira-text-dim); font-size: 0.85rem; font-weight: 500; }
        main { flex: 1; display: flex; flex-direction: column; overflow: hidden; min-height: 0; }
        main.scrollable { overflow-y: auto; }
      </style>
      <div class="shell">
        <header>
          <a class="logo" href="/_ui/">
            <span class="logo-icon">${this.meta?.logo?`<img src="${p(this.meta.logo)}" alt="logo">`:q}</span>
            <span class="logo-text">home</span>
          </a>
          ${this.meta&&this.meta.mode!=="index"?`
            <span class="sep">/</span>
            <span class="title">${p(this.meta.title||this.meta.name||"")}</span>
          `:""}
        </header>
        <main id="content" class="${this.meta?.mode!=="chat"?"scrollable":""}"></main>
      </div>
    `,_e(e.host,{theme:this.meta?.theme,accent:this.meta?.accent});let t=e.getElementById("content");if(!this.meta){t.innerHTML='<p style="padding:2rem;color:var(--haira-muted)">No workflow metadata found.</p>';return}switch(this.meta.mode){case"index":{let r=document.createElement("haira-index");r.setAttribute("data-meta",JSON.stringify(this.meta)),t.appendChild(r);break}case"form":{let r=document.createElement("haira-form");r.setAttribute("data-meta",JSON.stringify(this.meta)),t.appendChild(r);break}case"chat":{let r=document.createElement("haira-chat");r.setAttribute("data-meta",JSON.stringify(this.meta)),t.appendChild(r);break}}}}class de extends HTMLElement{_fileValue=null;connectedCallback(){let e=this.getAttribute("name")||"",t=this.getAttribute("type")||"string",r=this.attachShadow({mode:"open"});if(r.innerHTML=`
      <style>
        ${$}
        :host { display: block; margin-bottom: 1rem; }
        .field-label { display: block; font-weight: 500; font-size: 0.82rem; color: var(--haira-text-dim); margin-bottom: 0.4rem; }
        .field-sublabel { font-weight: 400; color: var(--haira-muted); font-size: 0.75rem; margin-left: 0.25rem; }
        input[type="text"], input[type="number"] {
          width: 100%; padding: 0.55rem 0.75rem; background: var(--haira-bg-input);
          border: 1px solid var(--haira-border); border-radius: var(--haira-radius-sm);
          color: var(--haira-text); font-size: 0.88rem; font-family: var(--haira-font);
          outline: none; transition: border-color 0.15s, box-shadow 0.15s;
        }
        input[type="text"]:focus, input[type="number"]:focus {
          border-color: var(--haira-accent); box-shadow: 0 0 0 3px rgba(232, 163, 23, 0.08);
        }
        input[type="text"]::placeholder, input[type="number"]::placeholder { color: var(--haira-muted); opacity: 0.6; }
        .toggle-row { display: flex; align-items: center; justify-content: space-between; padding: 0.5rem 0; }
        .toggle-label { font-size: 0.88rem; font-weight: 500; color: var(--haira-text-dim); }
        .toggle { position: relative; width: 40px; height: 22px; flex-shrink: 0; }
        .toggle input { opacity: 0; width: 0; height: 0; position: absolute; }
        .toggle-track {
          position: absolute; inset: 0; background: var(--haira-bg-elevated);
          border: 1px solid var(--haira-border); border-radius: 11px;
          cursor: pointer; transition: background 0.2s, border-color 0.2s;
        }
        .toggle-track::after {
          content: ""; position: absolute; top: 2px; left: 2px;
          width: 16px; height: 16px; background: var(--haira-muted);
          border-radius: 50%; transition: transform 0.2s ease, background 0.2s;
        }
        .toggle input:checked + .toggle-track { background: rgba(232, 163, 23, 0.15); border-color: var(--haira-accent); }
        .toggle input:checked + .toggle-track::after { transform: translateX(18px); background: var(--haira-accent); }
        .toggle input:focus-visible + .toggle-track { box-shadow: 0 0 0 3px rgba(232, 163, 23, 0.15); }
        .drop-zone {
          position: relative; border: 1.5px dashed var(--haira-border);
          border-radius: var(--haira-radius); padding: 1.25rem 1rem;
          text-align: center; cursor: pointer; transition: all 0.2s; background: var(--haira-bg-input);
        }
        .drop-zone:hover, .drop-zone.dragover { border-color: var(--haira-accent); background: rgba(232, 163, 23, 0.03); }
        .drop-zone.has-file { border-style: solid; border-color: var(--haira-success); background: rgba(34, 197, 94, 0.03); }
        .drop-icon { margin-bottom: 0.35rem; color: var(--haira-muted); }
        .drop-zone.has-file .drop-icon { color: var(--haira-success); }
        .drop-text { font-size: 0.82rem; color: var(--haira-muted); line-height: 1.4; }
        .drop-text strong { color: var(--haira-accent); cursor: pointer; }
        .drop-zone.has-file .drop-text { color: var(--haira-text-dim); }
        .drop-text .filename { display: block; font-family: var(--haira-mono); font-size: 0.8rem; color: var(--haira-text); margin-top: 0.2rem; word-break: break-all; }
        .drop-text .filesize { font-size: 0.72rem; color: var(--haira-muted); }
        .drop-zone input[type="file"] { position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer; }
        .clear-btn {
          display: none; position: absolute; top: 0.5rem; right: 0.5rem;
          background: var(--haira-bg-elevated); border: 1px solid var(--haira-border);
          border-radius: 4px; color: var(--haira-muted); font-size: 0.7rem;
          padding: 0.15rem 0.4rem; cursor: pointer; transition: all 0.15s;
        }
        .clear-btn:hover { color: var(--haira-error); border-color: var(--haira-error); }
        .drop-zone.has-file .clear-btn { display: block; }
      </style>
      ${this.renderInput(e,t)}
    `,t==="file")this.setupFileDrop(r)}renderInput(e,t){switch(t){case"bool":return`
          <div class="toggle-row">
            <span class="toggle-label">${p(e)}</span>
            <label class="toggle">
              <input type="checkbox" id="f-${e}" name="${e}">
              <span class="toggle-track"></span>
            </label>
          </div>`;case"file":return`
          <label class="field-label">${p(e)}</label>
          <div class="drop-zone" id="drop-zone">
            <input type="file" id="f-${e}" name="${e}">
            <div class="drop-icon">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none"><path d="M12 16V4m0 0l-4 4m4-4l4 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><path d="M4 17v2a2 2 0 002 2h12a2 2 0 002-2v-2" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
            </div>
            <div class="drop-text" id="drop-text">
              Drop file here or <strong>browse</strong>
            </div>
            <button class="clear-btn" id="clear-btn" type="button">Clear</button>
          </div>`;case"int":return`
          <label class="field-label" for="f-${e}">${p(e)} <span class="field-sublabel">integer</span></label>
          <input type="number" id="f-${e}" name="${e}" step="1" placeholder="0">`;case"float":return`
          <label class="field-label" for="f-${e}">${p(e)} <span class="field-sublabel">number</span></label>
          <input type="number" id="f-${e}" name="${e}" step="any" placeholder="0.0">`;default:return`
          <label class="field-label" for="f-${e}">${p(e)}</label>
          <input type="text" id="f-${e}" name="${e}" placeholder="Enter value...">`}}setupFileDrop(e){let t=e.getElementById("drop-zone"),r=e.querySelector("input[type=file]"),i=e.getElementById("drop-text"),o=e.getElementById("clear-btn");if(!t||!r||!i||!o)return;let a=(s)=>{this._fileValue=s,t.classList.add("has-file");let l=s.size<1024?`${s.size} B`:s.size<1048576?`${(s.size/1024).toFixed(1)} KB`:`${(s.size/1048576).toFixed(1)} MB`;i.innerHTML=`<span class="filename">${p(s.name)}</span><span class="filesize">${l}</span>`},n=()=>{this._fileValue=null,r.value="",t.classList.remove("has-file"),i.innerHTML="Drop file here or <strong>browse</strong>"};r.addEventListener("change",()=>{if(r.files?.[0])a(r.files[0])}),o.addEventListener("click",(s)=>{s.stopPropagation(),n()}),t.addEventListener("dragover",(s)=>{s.preventDefault(),t.classList.add("dragover")}),t.addEventListener("dragleave",()=>{t.classList.remove("dragover")}),t.addEventListener("drop",(s)=>{s.preventDefault(),t.classList.remove("dragover");let l=s.dataTransfer?.files?.[0];if(l){a(l);let m=new DataTransfer;m.items.add(l),r.files=m.files}})}getValue(){let e=this.getAttribute("name")||"",t=this.getAttribute("type")||"string";if(t==="bool"){let i=this.shadowRoot.querySelector("input[type=checkbox]");return{name:e,value:i.checked,type:t}}if(t==="file")return{name:e,value:this._fileValue||this.shadowRoot.querySelector("input[type=file]")?.files?.[0]||null,type:t};let r=this.shadowRoot.querySelector("input");if(t==="int"||t==="float")return{name:e,value:r.value?Number(r.value):"",type:t};return{name:e,value:r.value,type:t}}}class ce extends HTMLElement{rawText="";connectedCallback(){let e=this.attachShadow({mode:"open"});e.innerHTML=`
      <style>
        ${$}
        :host { display: none; margin-top: 0.75rem; }
        :host([visible]) { display: block; animation: fadeSlideUp 0.25s ease-out; }
        .card {
          background: var(--haira-bg-card); border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius); overflow: hidden;
        }
        .header {
          display: flex; align-items: center; justify-content: space-between;
          padding: 0.6rem 0.85rem; border-bottom: 1px solid var(--haira-border);
        }
        .header-left { display: flex; align-items: center; gap: 0.4rem; font-weight: 600; font-size: 0.78rem; color: var(--haira-muted); }
        .dot { width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
        .dot.success { background: var(--haira-success); }
        .dot.error { background: var(--haira-error); }
        .copy-btn {
          background: none; border: 1px solid transparent; border-radius: 4px;
          padding: 0.25rem; cursor: pointer; color: var(--haira-muted);
          display: flex; align-items: center; justify-content: center; transition: all 0.15s;
        }
        .copy-btn:hover { color: var(--haira-accent); border-color: var(--haira-border); background: var(--haira-bg-elevated); }
        .body {
          padding: 0.85rem; font-size: 0.82rem; line-height: 1.6;
          color: var(--haira-text-dim); max-height: 600px; overflow-y: auto; ${R}
        }
        .body.rich { white-space: normal; word-break: break-word; font-family: var(--haira-font); }
        .body.raw { white-space: pre-wrap; word-break: break-word; font-family: var(--haira-mono); font-size: 0.8rem; }
        .result-section { margin-bottom: 0.75rem; }
        .result-section:last-child { margin-bottom: 0; }
        .section-label {
          font-size: 0.68rem; font-weight: 700; text-transform: uppercase;
          letter-spacing: 0.05em; color: var(--haira-muted); margin-bottom: 0.3rem;
        }
        .section-label.error { color: var(--haira-error); }
        .section-value { color: var(--haira-text); line-height: 1.55; }
        .section-value ul { margin: 0.25rem 0 0 0; padding-left: 1.25rem; }
        .section-value li { margin-bottom: 0.15rem; font-family: var(--haira-mono); font-size: 0.78rem; color: var(--haira-text-dim); }
        .code-block {
          background: var(--haira-bg); border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius-sm); padding: 0.65rem 0.85rem; margin-top: 0.35rem;
          font-family: var(--haira-mono); font-size: 0.75rem; line-height: 1.5;
          white-space: pre-wrap; word-break: break-all; overflow-x: auto; color: var(--haira-text);
        }
        .code-lang { font-size: 0.62rem; text-transform: uppercase; color: var(--haira-muted); letter-spacing: 0.04em; margin-bottom: 0.2rem; font-weight: 600; }
        .result-kv { display: flex; gap: 0.5rem; padding: 0.2rem 0; }
        .result-kv .kv-key { font-size: 0.72rem; font-weight: 600; color: var(--haira-muted); min-width: 60px; flex-shrink: 0; }
        .result-kv .kv-val { color: var(--haira-text); font-size: 0.82rem; }
      </style>
      <div class="card">
        <div class="header">
          <div class="header-left">
            <span class="dot" id="dot"></span>
            <span id="label">Result</span>
          </div>
          <button class="copy-btn" id="copy-btn" title="Copy to clipboard">${h.copy}</button>
        </div>
        <div class="body raw" id="body"></div>
      </div>
    `,e.getElementById("copy-btn").addEventListener("click",()=>this.copyResult())}show(e,t){this.setAttribute("visible","");let r=this.shadowRoot.getElementById("body"),i=this.shadowRoot.getElementById("dot"),o=this.shadowRoot.getElementById("label");i.className=`dot ${t?"error":"success"}`,o.textContent=t?"Error":"Result";let a=e;if(typeof e==="object"&&e!==null&&typeof a.message==="string"&&a.message.length>0){this.rawText=a.message,r.className="body rich",r.innerHTML=this.renderMessage(a.message,a.status);return}if(typeof e==="object"&&e!==null&&!Array.isArray(e)){let s=Object.keys(a);if(s.length>0&&s.length<=10&&s.every((l)=>typeof a[l]!=="object"||a[l]===null)){this.rawText=JSON.stringify(e,null,2),r.className="body rich",r.innerHTML=s.map((l)=>`<div class="result-kv"><span class="kv-key">${p(l)}</span><span class="kv-val">${p(String(a[l]??""))}</span></div>`).join("");return}}let n;if(typeof e==="string")n=e;else n=JSON.stringify(e,null,2);this.rawText=n,r.className="body raw",r.textContent=n}hide(){this.removeAttribute("visible")}renderMessage(e,t){let r=e.split(`
`),i=[],o=0;while(o<r.length){let a=r[o],n=a.match(/^```(\w*)$/);if(n){let l=n[1]||"",m=[];o++;while(o<r.length&&!r[o].startsWith("```"))m.push(r[o]),o++;o++,i.push({type:"code",lang:l,content:m.join(`
`)});continue}let s=a.match(/^([A-Z][A-Z _]{2,}):(.*)$/);if(s){let l=s[1].trim(),m=s[2].trim(),u=m?[m]:[];o++;while(o<r.length){let g=r[o];if(g.match(/^[A-Z][A-Z _]{2,}:/)||g.startsWith("```"))break;u.push(g),o++}let d=u.join(`
`).trim(),c=d.split(`
`);if(c.length>1&&c.every((g)=>g.startsWith("- ")||g.trim()===""))i.push({type:"list",label:l,content:d});else i.push({type:"heading",label:l,content:d});continue}if(a.trim())i.push({type:"text",content:a});o++}if(i.length===0)return`<div class="section-value">${p(e)}</div>`;return i.map((a)=>{switch(a.type){case"heading":return`<div class="result-section">
              <div class="section-label${t==="error"&&a.label?.includes("CAUSE")?" error":""}">${p(a.label||"")}</div>
              <div class="section-value">${p(a.content)}</div>
            </div>`;case"list":return`<div class="result-section">
              <div class="section-label">${p(a.label||"")}</div>
              <div class="section-value"><ul>${a.content.split(`
`).filter((n)=>n.startsWith("- ")).map((n)=>`<li>${p(n.slice(2))}</li>`).join("")}</ul></div>
            </div>`;case"code":return`<div class="result-section">
              ${a.lang?`<div class="code-lang">${p(a.lang)}</div>`:""}
              <div class="code-block">${p(a.content)}</div>
            </div>`;case"text":return`<div class="section-value">${p(a.content)}</div>`;default:return""}}).join("")}async copyResult(){let e=this.shadowRoot?.getElementById("copy-btn");if(!e)return;try{await navigator.clipboard.writeText(this.rawText),e.innerHTML=h.copyDone,setTimeout(()=>{e.innerHTML=h.copy},1500)}catch{}}}class pe extends HTMLElement{_status="pending";_duration;_timerInterval=null;_timerStart=0;_expanded=!1;_logCount=0;_hasError=!1;connectedCallback(){this.renderStep()}disconnectedCallback(){this.clearTimer()}renderStep(){let e=this.getAttribute("name")||"",t=this.getAttribute("index")||"0",r=this.attachShadow({mode:"open"});r.innerHTML=`
      <style>
        ${$}
        :host { display: block; position: relative; }
        .step-header {
          display: flex; align-items: center; gap: 0.6rem;
          padding: 0.5rem 0.65rem; border-radius: var(--haira-radius-sm);
          cursor: pointer; user-select: none; transition: background 0.15s; position: relative;
        }
        .step-header:hover { background: rgba(255, 255, 255, 0.03); }
        .chevron {
          flex-shrink: 0; width: 16px; height: 16px; display: flex; align-items: center;
          justify-content: center; color: var(--haira-muted); transition: transform 0.2s ease, color 0.2s; opacity: 0;
        }
        .has-logs .chevron { opacity: 1; }
        .expanded .chevron { transform: rotate(90deg); }
        .status-icon {
          flex-shrink: 0; width: 22px; height: 22px; border-radius: 50%;
          display: flex; align-items: center; justify-content: center; transition: all 0.25s ease;
        }
        .pending .status-icon { border: 1.5px dashed var(--haira-muted); color: var(--haira-muted); }
        .running .status-icon { border: 1.5px solid var(--haira-accent); color: var(--haira-accent); background: rgba(232, 163, 23, 0.1); }
        .done .status-icon { background: var(--haira-success); color: #fff; }
        .failed .status-icon { background: var(--haira-error); color: #fff; }
        .retrying .status-icon {
          border: 1.5px solid var(--haira-accent); color: var(--haira-accent);
          background: rgba(232, 163, 23, 0.1); animation: pulse 1.5s ease-in-out infinite;
        }
        .skipped .status-icon { border: 1.5px dashed var(--haira-muted); color: var(--haira-muted); opacity: 0.5; }
        .step-num { font-size: 0.65rem; font-weight: 600; }
        .step-name {
          flex: 1; font-size: 0.85rem; font-weight: 500; color: var(--haira-muted);
          overflow: hidden; text-overflow: ellipsis; white-space: nowrap; transition: color 0.2s;
        }
        .running .step-name { color: var(--haira-text); font-weight: 600; }
        .done .step-name { color: var(--haira-text-dim); }
        .failed .step-name { color: var(--haira-text); }
        .retrying .step-name { color: var(--haira-text); }
        .log-count {
          font-size: 0.7rem; color: var(--haira-muted); padding: 0.1rem 0.4rem;
          border-radius: 10px; background: rgba(255, 255, 255, 0.04);
          font-family: var(--haira-mono); display: none;
        }
        .has-logs .log-count { display: inline-block; }
        .has-error .log-count { color: var(--haira-error); background: rgba(239, 68, 68, 0.1); }
        .timer {
          flex-shrink: 0; font-size: 0.75rem; font-family: var(--haira-mono);
          color: var(--haira-muted); min-width: 36px; text-align: right; transition: color 0.2s;
        }
        .running .timer { color: var(--haira-accent-light); }
        .done .timer { color: var(--haira-success); }
        .failed .timer { color: var(--haira-error); }
        .logs-wrapper {
          overflow: hidden; max-height: 0; opacity: 0;
          transition: max-height 0.25s ease, opacity 0.2s ease; margin-left: 2.55rem;
        }
        .logs-wrapper.open { max-height: 600px; opacity: 1; overflow-y: auto; ${R} }
        .logs-inner { padding: 0.25rem 0 0.5rem 0; border-left: 1px solid rgba(63, 63, 70, 0.3); margin-left: 0.15rem; }
        .log-entry {
          display: flex; align-items: flex-start; gap: 0.5rem;
          font-size: 0.78rem; font-family: var(--haira-mono); line-height: 1.5;
          padding: 0.15rem 0 0.15rem 0.85rem; animation: fadeIn 0.15s ease-out both;
        }
        .log-badge {
          flex-shrink: 0; font-size: 0.6rem; font-weight: 700; text-transform: uppercase;
          padding: 0.08rem 0.35rem; border-radius: 3px; letter-spacing: 0.04em; margin-top: 0.12rem;
        }
        .log-badge.info { background: rgba(59, 130, 246, 0.12); color: var(--haira-info); }
        .log-badge.warn { background: rgba(234, 179, 8, 0.12); color: var(--haira-warn); }
        .log-badge.error { background: rgba(239, 68, 68, 0.12); color: var(--haira-error); }
        .log-msg { flex: 1; word-break: break-word; white-space: pre-wrap; color: var(--haira-text-dim); }
        .log-msg.warn { color: var(--haira-warn); }
        .log-msg.error { color: var(--haira-error); }
        .error-detail {
          margin: 0.25rem 0 0.5rem 0; padding: 0.5rem 0.75rem;
          font-size: 0.78rem; font-family: var(--haira-mono); color: var(--haira-error);
          background: rgba(239, 68, 68, 0.06); border: 1px solid rgba(239, 68, 68, 0.12);
          border-radius: var(--haira-radius-sm); margin-left: 2.55rem; line-height: 1.5;
          word-break: break-word; white-space: pre-wrap; display: none;
        }
        .error-detail.visible { display: block; animation: fadeIn 0.2s ease-out; }
      </style>
      <div class="step-header pending" id="header">
        <span class="chevron" id="chevron">${h.chevron}</span>
        <span class="status-icon" id="status-icon">
          <span class="step-num" id="step-num">${Number(t)+1}</span>
        </span>
        <span class="step-name">${p(e)}</span>
        <span class="log-count" id="log-count"></span>
        <span class="timer" id="timer"></span>
      </div>
      <div class="logs-wrapper" id="logs-wrapper">
        <div class="logs-inner" id="logs"></div>
      </div>
      <div class="error-detail" id="error-detail"></div>
    `,r.getElementById("header").addEventListener("click",()=>{if(this._logCount===0)return;this.toggleLogs()})}toggleLogs(e){let t=this.shadowRoot?.getElementById("logs-wrapper"),r=this.shadowRoot?.getElementById("header");if(!t||!r)return;if(e===!0)this._expanded=!0;else if(e===!1)this._expanded=!1;else this._expanded=!this._expanded;if(this._expanded)t.classList.add("open"),r.classList.add("expanded");else t.classList.remove("open"),r.classList.remove("expanded")}setStatus(e,t,r){this._status=e,this._duration=t;let i=this.shadowRoot?.getElementById("header"),o=this.shadowRoot?.getElementById("status-icon"),a=this.shadowRoot?.getElementById("timer"),n=this.shadowRoot?.getElementById("error-detail");if(!i||!o||!a)return;let s=[];if(this._logCount>0)s.push("has-logs");if(this._hasError)s.push("has-error");if(this._expanded)s.push("expanded");switch(i.className=`step-header ${e} ${s.join(" ")}`,e){case"pending":o.innerHTML=`<span class="step-num">${this.getIndex()}</span>`,this.clearTimer(),a.textContent="",n?.classList.remove("visible");break;case"running":o.innerHTML=h.spinner,this.startTimer(a),n?.classList.remove("visible");break;case"done":if(o.innerHTML=h.check,this.clearTimer(),a.textContent=this.formatDuration(t),n?.classList.remove("visible"),!this._hasError)this.toggleLogs(!1);break;case"failed":if(o.innerHTML=h.x,this.clearTimer(),a.textContent=this.formatDuration(t),r&&n)n.textContent=r,n.classList.add("visible");if(this._logCount>0)this.toggleLogs(!0);break;case"retrying":o.innerHTML=h.retry,n?.classList.remove("visible");break;case"skipped":o.innerHTML=`<span class="step-num">${this.getIndex()}</span>`,this.clearTimer(),a.textContent="skipped",n?.classList.remove("visible");break}}addLog(e,t){let r=this.shadowRoot?.getElementById("logs"),i=this.shadowRoot?.getElementById("header"),o=this.shadowRoot?.getElementById("log-count");if(!r||!i||!o)return;if(this._logCount++,i.classList.add("has-logs"),o.textContent=String(this._logCount),e==="error")this._hasError=!0,i.classList.add("has-error"),this.toggleLogs(!0);let a=200,n=t.length>a?t.slice(0,a)+"...":t,s=document.createElement("div");if(s.className="log-entry",s.innerHTML=`<span class="log-badge ${e}">${e}</span><span class="log-msg ${e}">${p(n)}</span>`,r.appendChild(s),this._status==="running"&&!this._expanded)this.toggleLogs(!0);let l=this.shadowRoot?.getElementById("logs-wrapper");if(l&&this._expanded)l.scrollTop=l.scrollHeight}clearLogs(){let e=this.shadowRoot?.getElementById("logs"),t=this.shadowRoot?.getElementById("header"),r=this.shadowRoot?.getElementById("error-detail");if(!e)return;e.innerHTML="",this._logCount=0,this._hasError=!1,this._expanded=!1,t?.classList.remove("has-logs","has-error","expanded"),r?.classList.remove("visible"),this.toggleLogs(!1);let i=this.shadowRoot?.getElementById("log-count");if(i)i.textContent=""}getIndex(){let e=this.getAttribute("index");return e!==null?String(Number(e)+1):""}startTimer(e){this.clearTimer(),this._timerStart=performance.now(),e.textContent="0.0s",this._timerInterval=setInterval(()=>{let t=(performance.now()-this._timerStart)/1000;e.textContent=`${t.toFixed(1)}s`},100)}clearTimer(){if(this._timerInterval!==null)clearInterval(this._timerInterval),this._timerInterval=null}formatDuration(e){if(e===void 0)return"";return e<1000?`${e}ms`:`${(e/1000).toFixed(1)}s`}get status(){return this._status}get duration(){return this._duration}}class he extends HTMLElement{steps=[];stepElements=[];stepStatuses=[];totalDuration=0;connectedCallback(){let e=this.attachShadow({mode:"open"});e.innerHTML=`
      <style>
        ${$}
        :host { display: none; margin-top: 1.25rem; }
        :host([visible]) { display: block; animation: fadeIn 0.2s ease-out; }
        .header {
          display: flex; align-items: center; gap: 0.5rem;
          padding: 0 0.25rem 0.6rem; font-size: 0.72rem; font-weight: 600;
          color: var(--haira-muted); text-transform: uppercase; letter-spacing: 0.06em;
        }
        .header-line { flex: 1; height: 1px; background: var(--haira-border); }
        .pipeline {
          display: flex; flex-direction: column; gap: 1px;
          background: var(--haira-bg-card); border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius); padding: 0.35rem; overflow: hidden;
        }
        .summary {
          display: none; padding: 0.75rem 1rem; font-size: 0.78rem;
          color: var(--haira-muted); border-top: 1px solid var(--haira-border);
          margin-top: 0.5rem; border-radius: var(--haira-radius-sm);
          background: var(--haira-bg-card); text-align: center;
        }
        .summary.visible { display: block; animation: fadeIn 0.3s ease-out; }
        .summary .count { color: var(--haira-text-dim); font-weight: 500; }
        .summary .time { color: var(--haira-accent); font-family: var(--haira-mono); font-weight: 600; }
        .summary .failed-count { color: var(--haira-error); }
      </style>
      <div class="header">
        <span>Pipeline</span>
        <span class="header-line"></span>
      </div>
      <div class="pipeline" id="pipeline"></div>
      <div class="summary" id="summary"></div>
    `}setSteps(e){this.steps=e,this.stepElements=[],this.stepStatuses=e.map(()=>"pending"),this.totalDuration=0;let t=this.shadowRoot?.getElementById("pipeline"),r=this.shadowRoot?.getElementById("summary");if(!t)return;if(t.innerHTML="",r)r.classList.remove("visible"),r.textContent="";e.forEach((i,o)=>{let a=document.createElement("haira-step");a.setAttribute("name",i),a.setAttribute("index",String(o)),t.appendChild(a),this.stepElements.push(a)})}updateStep(e){let t=this.steps.indexOf(e.name);if(t===-1)return;let r=this.stepElements[t];if(!r)return;if(e.status==="log"&&e.log){r.addLog(e.log.level,e.log.message);return}let i;switch(e.status){case"start":i="running";break;case"end":if(i="done",e.duration_ms)this.totalDuration+=e.duration_ms;break;case"failed":i="failed";break;case"retry":i="retrying";break;default:return}this.stepStatuses[t]=i,r.setStatus(i,e.duration_ms,e.error),this.checkCompletion()}checkCompletion(){if(!this.stepStatuses.every((s)=>s==="done"||s==="failed"||s==="skipped"))return;let t=this.shadowRoot?.getElementById("summary");if(!t)return;let r=this.stepStatuses.filter((s)=>s==="done").length,i=this.stepStatuses.filter((s)=>s==="failed").length,o=this.stepStatuses.filter((s)=>s==="skipped").length,a=(this.totalDuration/1000).toFixed(1),n=`<span class="count">${r}/${this.steps.length} steps completed`;if(i>0)n+=` &middot; <span class="failed-count">${i} failed</span>`;if(o>0)n+=` &middot; ${o} skipped`;n+=`</span> &middot; <span class="time">${a}s total</span>`,t.innerHTML=n,t.classList.add("visible")}finalize(){for(let e=0;e<this.stepStatuses.length;e++){let t=this.stepStatuses[e];if(t==="running")this.stepStatuses[e]="done",this.stepElements[e].setStatus("done");else if(t==="pending")this.stepStatuses[e]="skipped",this.stepElements[e].setStatus("skipped")}this.checkCompletion()}reset(){this.totalDuration=0,this.stepStatuses=this.steps.map(()=>"pending");for(let t of this.stepElements)t.clearLogs(),t.setStatus("pending");let e=this.shadowRoot?.getElementById("summary");if(e)e.classList.remove("visible"),e.innerHTML=""}show(){this.setAttribute("visible","")}hide(){this.removeAttribute("visible")}}class me extends HTMLElement{connectedCallback(){this.renderMessage()}renderMessage(){let e=this.getAttribute("role")||"user",t=this.getAttribute("content")||"",r=this.getAttribute("file")||"",i=this.getAttribute("avatar")||"H",o=this.attachShadow({mode:"open"});o.innerHTML=`
      <style>
        ${$}
        :host { display: block; animation: fadeSlideUp 0.2s ease-out; }
        .row { display: flex; gap: 0.6rem; align-items: flex-start; }
        .row.user { justify-content: flex-end; }
        .row.assistant { justify-content: flex-start; }
        .avatar {
          width: 26px; height: 26px; border-radius: 50%;
          display: flex; align-items: center; justify-content: center;
          flex-shrink: 0; font-size: 0.65rem; font-weight: 700; margin-top: 2px;
        }
        .avatar.assistant {
          background: var(--haira-accent-dim); border: 1px solid rgba(232, 163, 23, 0.2);
          color: var(--haira-accent);
        }
        .avatar.assistant img { width: 100%; height: 100%; border-radius: 50%; object-fit: cover; }
        .avatar.user { display: none; }
        .bubble {
          padding: 0.7rem 0.9rem; border-radius: 12px; line-height: 1.6;
          font-size: 0.88rem; word-wrap: break-word; min-width: 40px;
        }
        .bubble.user {
          background: var(--haira-bg-elevated); border: 1px solid var(--haira-border);
          color: var(--haira-text); border-bottom-right-radius: 4px; max-width: 85%;
        }
        .bubble.assistant { background: transparent; color: var(--haira-text); padding: 0.4rem 0; flex: 1; }
        .file-tag {
          display: inline-flex; align-items: center; gap: 0.3rem;
          background: rgba(0,0,0,0.12); padding: 0.2rem 0.5rem; border-radius: 5px;
          font-size: 0.75rem; margin-bottom: 0.3rem; font-weight: 500;
        }
        .file-tag svg { opacity: 0.7; }
        .bubble.assistant .code-wrapper { position: relative; margin: 0.5rem 0; }
        .bubble.assistant pre {
          background: var(--haira-bg); border: 1px solid var(--haira-border);
          padding: 0.6rem 0.75rem; border-radius: 6px; overflow-x: auto;
          font-size: 0.78rem; font-family: var(--haira-mono); line-height: 1.5; margin: 0;
        }
        .bubble.assistant code {
          background: var(--haira-bg-elevated); border: 1px solid var(--haira-border);
          padding: 0.1rem 0.3rem; border-radius: 3px; font-size: 0.8rem;
          font-family: var(--haira-mono); color: var(--haira-accent-light);
        }
        .bubble.assistant pre code { background: none; border: none; padding: 0; color: var(--haira-text); }
        .bubble.assistant strong { font-weight: 700; }
        .bubble.assistant em { font-style: italic; color: var(--haira-text-dim); }
        .bubble.assistant p { margin: 0.3rem 0; }
        .bubble.assistant p:first-child { margin-top: 0; }
        .bubble.assistant p:last-child { margin-bottom: 0; }
        .bubble.assistant ul, .bubble.assistant ol { margin: 0.35rem 0; padding-left: 1.3rem; }
        .bubble.assistant li { margin: 0.2rem 0; }
        .bubble.assistant h1, .bubble.assistant h2, .bubble.assistant h3 {
          font-size: 0.9rem; font-weight: 700; margin: 0.6rem 0 0.25rem; color: var(--haira-text);
        }
        .bubble.assistant h1:first-child, .bubble.assistant h2:first-child, .bubble.assistant h3:first-child { margin-top: 0; }
        .bubble.assistant hr { border: none; border-top: 1px solid var(--haira-border); margin: 0.5rem 0; }
        .bubble.assistant a { color: var(--haira-accent); text-decoration: none; }
        .bubble.assistant a:hover { text-decoration: underline; }
        .bubble.assistant blockquote {
          border-left: 3px solid var(--haira-accent); margin: 0.4rem 0;
          padding: 0.2rem 0.6rem; color: var(--haira-text-dim);
        }
        .bubble.assistant table { border-collapse: collapse; width: 100%; margin: 0.5rem 0; font-size: 0.82rem; }
        .bubble.assistant th {
          text-align: left; padding: 0.4rem 0.6rem; border-bottom: 2px solid var(--haira-border);
          font-weight: 600; color: var(--haira-text); font-size: 0.78rem;
        }
        .bubble.assistant td { padding: 0.35rem 0.6rem; border-bottom: 1px solid var(--haira-border); color: var(--haira-text-dim); }
        .bubble.assistant tr:last-child td { border-bottom: none; }
        .bubble.assistant ol { margin: 0.35rem 0; padding-left: 1.3rem; }
        .bubble.assistant ol li { margin: 0.2rem 0; }
        .copy-code {
          position: absolute; top: 0.4rem; right: 0.4rem;
          background: var(--haira-bg-card); border: 1px solid var(--haira-border);
          border-radius: 4px; padding: 0.2rem; cursor: pointer;
          color: var(--haira-muted); display: flex; align-items: center;
          opacity: 0; transition: opacity 0.15s;
        }
        .code-wrapper:hover .copy-code { opacity: 1; }
        .copy-code:hover { color: var(--haira-accent); border-color: var(--haira-accent); }
      </style>
      <div class="row ${e}">
        ${e==="assistant"?`<div class="avatar assistant">${i.startsWith("http")?`<img src="${p(i)}" alt="">`:p(i)}</div>`:""}
        <div class="bubble ${e}" id="bubble"></div>
      </div>
    `;let a=o.getElementById("bubble");if(e==="assistant")a.innerHTML=this.renderMarkdown(t),this.attachCopyHandlers(o);else{let n="";if(r)n+=`<div class="file-tag">${h.file} ${p(r)}</div><br>`;if(t)n+=p(t);a.innerHTML=n}}updateContent(e){let t=this.shadowRoot?.getElementById("bubble");if(t)t.innerHTML=this.renderMarkdown(e),this.attachCopyHandlers(this.shadowRoot)}attachCopyHandlers(e){e.querySelectorAll(".copy-code").forEach((t)=>{t.addEventListener("click",async()=>{let i=t.closest(".code-wrapper")?.querySelector("code");if(!i)return;try{await navigator.clipboard.writeText(i.textContent||""),t.innerHTML=h.copyDone,setTimeout(()=>{t.innerHTML=h.copy},1500)}catch{}})})}renderMarkdown(e){let t=p(e);if(t=t.replace(/```(\w*)\n([\s\S]*?)```/g,(r,i,o)=>`<div class="code-wrapper"><pre><code>${o.trim()}</code></pre><button class="copy-code" title="Copy code">${h.copy}</button></div>`),t=t.replace(/((?:^|\n)\|.+\|(?:\n\|[-:| ]+\|)(?:\n\|.+\|)+)/g,(r)=>{let i=r.trim().split(`
`);if(i.length<2)return r;let o=i[0].split("|").filter((s)=>s.trim()),a=i.slice(2),n="<table><thead><tr>";for(let s of o)n+=`<th>${s.trim()}</th>`;n+="</tr></thead><tbody>";for(let s of a){let l=s.split("|").filter((m)=>m.trim());n+="<tr>";for(let m of l)n+=`<td>${m.trim()}</td>`;n+="</tr>"}return n+="</tbody></table>",n}),t=t.replace(/`([^`]+)`/g,"<code>$1</code>"),t=t.replace(/^### (.+)$/gm,"<h3>$1</h3>"),t=t.replace(/^## (.+)$/gm,"<h2>$1</h2>"),t=t.replace(/^# (.+)$/gm,"<h1>$1</h1>"),t=t.replace(/^---$/gm,"<hr>"),t=t.replace(/\*\*(.+?)\*\*/g,"<strong>$1</strong>"),t=t.replace(/(?<!\w)\*(.+?)\*(?!\w)/g,"<em>$1</em>"),t=t.replace(/\[([^\]]+)\]\(([^)]+)\)/g,'<a href="$2" target="_blank" rel="noopener">$1</a>'),t=t.replace(/((?:^|\n)\d+\. .+(?:\n\d+\. .+)*)/g,(r)=>{let i=r.trim().split(`
`),o="<ol>";for(let a of i)o+=`<li>${a.replace(/^\d+\.\s+/,"")}</li>`;return o+="</ol>",o}),t=t.replace(/((?:^|\n)- .+(?:\n- .+)*)/g,(r)=>{let i=r.trim().split(`
`),o="<ul>";for(let a of i)o+=`<li>${a.replace(/^-\s+/,"")}</li>`;return o+="</ul>",o}),t=t.replace(/\n\n/g,"</p><p>"),t=t.replace(/\n/g,"<br>"),!t.startsWith("<"))t=`<p>${t}</p>`;return t}}async function Ie(e,t,r){let i=new TextDecoder,o="",a="";try{while(!0){let{done:n,value:s}=await e.read();if(n)break;o+=i.decode(s,{stream:!0});let l=o.split(`
`);o=l.pop();for(let m of l){let u=m.trim();if(u.startsWith("event:")){a=u.slice(6).trim();continue}if(!u.startsWith("data:"))continue;let d=u.slice(5).trim();if(d==="[DONE]"){t.onDone?.();return}try{let c=JSON.parse(d);switch(a){case"run_id":t.onRunId?.(c.run_id);break;case"step":t.onStep?.(c);break;case"result":t.onResult?.(c);break;case"error":t.onError?.(c.error||"Unknown error");break;case"delta":t.onDelta?.(c.delta);break;case"tool_start":t.onToolStart?.(c);break;case"tool_end":t.onToolEnd?.(c);break;case"tool_render":t.onToolRender?.(c);break;default:if(c.delta)t.onDelta?.(c.delta);break}}catch{}a=""}}t.onDone?.()}catch(n){if(r?.aborted)return;throw n}}async function X(e,t,r,i,o){let a={Accept:"text/event-stream"},n;if(i)n=i;else a["Content-Type"]="application/json",n=JSON.stringify(t);let s=await fetch(e,{method:"POST",headers:a,body:n,signal:o});if(!s.ok){let l=await s.text();r.onError?.(l||`HTTP ${s.status}`),r.onDone?.();return}await Ie(s.body.getReader(),r,o)}async function Be(e,t){let r=await fetch(e,{method:"GET",headers:{Accept:"text/event-stream"}});if(!r.ok){let i=await r.text();t.onError?.(i||`HTTP ${r.status}`),t.onDone?.();return}await Ie(r.body.getReader(),t)}async function Pe(e,t,r,i,o){let a;if(t==="GET"||t==="DELETE"){let m=new URLSearchParams(r),u=m.toString()?`${e}?${m}`:e;a={method:t},e=u}else if(i&&o)a={method:t,body:o};else a={method:t,headers:{"Content-Type":"application/json"},body:JSON.stringify(r)};let n=await fetch(e,a),s=await n.text(),l;try{l=JSON.parse(s)}catch{l=s}return{status:n.status,data:l}}class ue extends HTMLElement{meta;connectedCallback(){this.meta=JSON.parse(this.getAttribute("data-meta")||"{}"),this.renderForm()}renderForm(){let e=this.meta,t=this.attachShadow({mode:"open"});t.innerHTML=`
      <style>
        ${$}
        :host { display: block; padding: 2.5rem 1rem 3rem; }
        .layout { max-width: 960px; margin: 0 auto; width: 100%; }
        @media (min-width: 768px) { :host { padding: 2.5rem 2rem 3rem; } }
        .header { margin-bottom: 1.5rem; }
        h1 {
          font-size: 1.3rem; font-weight: 700; color: var(--haira-text);
          display: flex; align-items: center; gap: 0.6rem; margin-bottom: 0.25rem;
        }
        .method-badge {
          font-size: 0.6rem; font-weight: 700; padding: 0.15rem 0.45rem;
          border-radius: 4px; color: #fff; letter-spacing: 0.02em; flex-shrink: 0;
        }
        .desc { font-size: 0.85rem; color: var(--haira-muted); line-height: 1.45; }
        .path { font-family: var(--haira-mono); font-size: 0.78rem; color: var(--haira-muted); opacity: 0.7; margin-top: 0.15rem; }
        .card {
          background: var(--haira-bg-card); border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius); padding: 1.25rem; transition: opacity 0.2s;
        }
        .card.disabled { opacity: 0.45; pointer-events: none; }
        .submit-btn {
          width: 100%; padding: 0.65rem 1.5rem; border: none;
          background: var(--haira-accent); color: #0a0a0a;
          border-radius: var(--haira-radius-sm); font-size: 0.88rem; font-weight: 600;
          cursor: pointer; font-family: var(--haira-font); transition: all 0.15s;
          margin-top: 0.5rem; display: flex; align-items: center; justify-content: center; gap: 0.4rem;
        }
        .submit-btn:hover:not(:disabled) { background: var(--haira-accent-light); box-shadow: 0 2px 16px rgba(232, 163, 23, 0.2); }
        .submit-btn:active:not(:disabled) { transform: scale(0.99); }
        .submit-btn:disabled { opacity: 0.5; cursor: not-allowed; }
        .spinner {
          display: inline-block; width: 14px; height: 14px;
          border: 2px solid #0a0a0a; border-top-color: transparent;
          border-radius: 50%; animation: spin 0.6s linear infinite;
        }
        .output-area { margin-top: 0.5rem; }
      </style>
      <div class="layout">
        <div class="header">
          <h1>
            ${p(e.title||e.name)}
            <span class="method-badge" style="background:${D(e.method)}">${e.method}</span>
          </h1>
          ${e.description?`<p class="desc">${p(e.description)}</p>`:""}
          <p class="path">${p(e.path)}</p>
        </div>
        <div class="card" id="card">
          <form id="wf-form">
            <div id="fields"></div>
            <button type="submit" class="submit-btn" id="submit-btn">Run</button>
          </form>
        </div>
        <div class="output-area" id="output-area"></div>
      </div>
    `;let r=t.getElementById("fields");for(let u of e.params){let d=document.createElement("haira-field");d.setAttribute("name",u.Name),d.setAttribute("type",u.Type),r.appendChild(d)}let i=t.getElementById("output-area"),o=document.createElement("haira-pipeline");if(i.appendChild(o),e.steps&&e.steps.length>0)o.setSteps(e.steps);let a=document.createElement("haira-result");i.appendChild(a);let n=t.getElementById("wf-form"),s=t.getElementById("submit-btn"),l=t.getElementById("card"),m=new URLSearchParams(window.location.search).get("run");if(m)this.loadRun(m,o,a,l,s);n.addEventListener("submit",async(u)=>{u.preventDefault(),s.disabled=!0,s.innerHTML='<span class="spinner"></span>Running...',l.classList.add("disabled"),a.hide();let d=r.querySelectorAll("haira-field"),c={},f,g=!1;for(let E of d){let{name:C,value:H,type:V}=E.getValue();if(V==="file"&&H){if(g=!0,!f)f=new FormData;f.append(C,H)}else if(H!==""&&H!==null){if(c[C]=H,f)f.append(C,String(H))}}let v=()=>{s.disabled=!1,s.textContent="Run",l.classList.remove("disabled"),history.replaceState(null,"",window.location.pathname)};if(e.steps&&e.steps.length>0){o.reset(),o.show();let E;if(g&&f){for(let[C,H]of Object.entries(c))if(!f.has(C))f.append(C,String(H));E=f}await X(e.path,c,{onRunId:(C)=>{history.replaceState(null,"",`${window.location.pathname}?run=${C}`)},onStep:(C)=>{o.updateStep(C)},onResult:(C)=>{a.show(C,!1)},onError:(C)=>{a.show({error:C},!0)},onDone:()=>{o.finalize(),v()}},E)}else{try{let E=await Pe(e.path,e.method,c,g,f);a.show(E.data,E.status>=400)}catch(E){a.show({error:E.message},!0)}v()}})}async loadRun(e,t,r,i,o){let a;try{let n=await fetch(`/_api/runs/${e}`);if(!n.ok)return;a=await n.json()}catch{return}t.show();for(let n of a.steps)t.updateStep(n);if(a.status==="completed"&&a.result)r.show(a.result,!1),t.finalize();else if(a.status==="failed"){if(a.error)r.show({error:a.error},!0);t.finalize()}else if(a.status==="running")o.disabled=!0,o.innerHTML='<span class="spinner"></span>Running...',i.classList.add("disabled"),await Be(`/_api/runs/stream/${e}`,{onStep:(n)=>{t.updateStep(n)},onResult:(n)=>{r.show(n,!1)},onError:(n)=>{r.show({error:n},!0)},onDone:()=>{t.finalize(),o.disabled=!1,o.textContent="Run",i.classList.remove("disabled"),history.replaceState(null,"",window.location.pathname)}})}}class fe extends HTMLElement{connectedCallback(){let t=JSON.parse(this.getAttribute("data-meta")||"{}").workflows||[],r=this.attachShadow({mode:"open"});r.innerHTML=`
      <style>
        ${$}
        :host { display: flex; justify-content: center; padding: 2.5rem 1rem; }
        .container { max-width: 960px; width: 100%; }
        @media (min-width: 768px) { :host { padding: 2.5rem 2rem; } }
        h1 { font-size: 1.3rem; font-weight: 700; color: var(--haira-text); margin-bottom: 1.25rem; }
        .wf {
          display: flex; align-items: center; justify-content: space-between;
          background: var(--haira-bg-card); border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius); padding: 0.85rem 1rem; margin-bottom: 0.5rem;
          text-decoration: none; color: var(--haira-text); transition: all 0.15s;
          animation: fadeSlideUp 0.3s ease-out both;
        }
        .wf:hover { border-color: rgba(232, 163, 23, 0.3); background: var(--haira-bg-card-hover); }
        .wf-left { display: flex; align-items: center; gap: 0.6rem; min-width: 0; }
        .badge {
          font-size: 0.6rem; font-weight: 700; padding: 0.12rem 0.4rem;
          border-radius: 3px; color: #fff; flex-shrink: 0; letter-spacing: 0.02em;
        }
        .wf-info { min-width: 0; }
        .wf-name {
          font-weight: 600; font-size: 0.88rem; white-space: nowrap;
          overflow: hidden; text-overflow: ellipsis;
        }
        .wf-path { font-family: var(--haira-mono); font-size: 0.75rem; color: var(--haira-muted); margin-top: 0.1rem; }
        .wf-right { display: flex; align-items: center; flex-shrink: 0; margin-left: 0.75rem; }
        .type-pill {
          font-size: 0.65rem; font-weight: 600; padding: 0.12rem 0.5rem;
          border-radius: 10px; border: 1px solid; text-transform: lowercase;
        }
        .empty { text-align: center; padding: 3rem 1rem; animation: fadeIn 0.4s ease-out; }
        .empty-title { color: var(--haira-text-dim); font-size: 0.95rem; font-weight: 500; margin-bottom: 0.25rem; }
        .empty-sub { color: var(--haira-muted); font-size: 0.82rem; }
        .section-title { font-size: 1rem; font-weight: 700; color: var(--haira-text); margin: 2rem 0 0.75rem; }
        .run {
          display: flex; align-items: center; gap: 0.6rem;
          background: var(--haira-bg-card); border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius-sm); padding: 0.6rem 0.85rem; margin-bottom: 0.35rem;
          text-decoration: none; color: var(--haira-text); transition: all 0.15s;
          animation: fadeIn 0.25s ease-out both;
        }
        .run:hover { border-color: rgba(232, 163, 23, 0.3); background: var(--haira-bg-card-hover); }
        .run-dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
        .run-dot.completed { background: var(--haira-success); }
        .run-dot.failed { background: var(--haira-error); }
        .run-dot.running { background: var(--haira-accent); animation: pulse 1.5s ease-in-out infinite; }
        .run-name {
          flex: 1; font-size: 0.82rem; font-weight: 500;
          overflow: hidden; text-overflow: ellipsis; white-space: nowrap; min-width: 0;
        }
        .run-name .run-id { font-family: var(--haira-mono); font-size: 0.7rem; color: var(--haira-muted); margin-left: 0.4rem; }
        .run-time { font-size: 0.72rem; font-family: var(--haira-mono); color: var(--haira-muted); flex-shrink: 0; }
        .run-status {
          font-size: 0.62rem; font-weight: 600; text-transform: uppercase;
          letter-spacing: 0.03em; flex-shrink: 0; padding: 0.08rem 0.35rem; border-radius: 3px;
        }
        .run-status.completed { color: var(--haira-success); background: rgba(34, 197, 94, 0.1); }
        .run-status.failed { color: var(--haira-error); background: rgba(239, 68, 68, 0.1); }
        .run-status.running { color: var(--haira-accent); background: rgba(232, 163, 23, 0.1); }
        .chat {
          display: flex; align-items: center; gap: 0.6rem;
          background: var(--haira-bg-card); border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius-sm); padding: 0.6rem 0.85rem; margin-bottom: 0.35rem;
          text-decoration: none; color: var(--haira-text); transition: all 0.15s;
          animation: fadeIn 0.25s ease-out both;
        }
        .chat:hover { border-color: rgba(232, 163, 23, 0.3); background: var(--haira-bg-card-hover); }
        .chat-icon { color: var(--haira-accent); display: flex; flex-shrink: 0; opacity: 0.6; }
        .chat-title {
          flex: 1; font-size: 0.82rem; font-weight: 500;
          overflow: hidden; text-overflow: ellipsis; white-space: nowrap; min-width: 0;
        }
        .chat-wf { font-family: var(--haira-mono); font-size: 0.7rem; color: var(--haira-muted); margin-left: 0.4rem; }
        .chat-time { font-size: 0.72rem; font-family: var(--haira-mono); color: var(--haira-muted); flex-shrink: 0; }
        .chat-count {
          font-size: 0.62rem; color: var(--haira-muted); flex-shrink: 0;
          padding: 0.08rem 0.35rem; border-radius: 3px; background: var(--haira-bg-elevated);
        }
      </style>
      <div class="container">
        <h1>Workflows</h1>
        ${t.length===0?`<div class="empty">
              <div class="empty-title">No workflows registered</div>
              <div class="empty-sub">Define a workflow in your .haira file to get started</div>
            </div>`:t.map((i,o)=>`
            <a class="wf" href="/_ui${i.path}" style="animation-delay:${o*50}ms">
              <div class="wf-left">
                <span class="badge" style="background:${D(i.method)}">${i.method}</span>
                <div class="wf-info">
                  <div class="wf-name">${p(i.title||i.name)}</div>
                  <div class="wf-path">${p(i.path)}</div>
                </div>
              </div>
              <div class="wf-right">
                <span class="type-pill" style="color:${Y(i.uiType)};border-color:${Y(i.uiType)}30">${i.uiType}</span>
              </div>
            </a>
          `).join("")}
        <div id="chats-section"></div>
        <div id="runs-section"></div>
      </div>
    `,this.loadChats(r),this.loadRuns(r)}async loadChats(e){let t=e.getElementById("chats-section");if(!t)return;try{let r=await fetch("/_api/chats");if(!r.ok)return;let i=await r.json();if(!i||i.length===0)return;let o='<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z"/></svg>';t.innerHTML=`
        <h2 class="section-title">Recent Chats</h2>
        ${i.map((a,n)=>`
          <a class="chat" href="/_ui${p(a.workflow_path)}?session=${p(a.id)}" style="animation-delay:${n*30}ms">
            <span class="chat-icon">${o}</span>
            <span class="chat-title">${p(a.title||"New chat")}<span class="chat-wf">${p(a.workflow_name)}</span></span>
            <span class="chat-time">${this.relativeTime(a.updated_at)}</span>
            <span class="chat-count">${a.message_count} msg</span>
          </a>
        `).join("")}
      `}catch{}}async loadRuns(e){let t=e.getElementById("runs-section");if(!t)return;try{let r=await fetch("/_api/runs");if(!r.ok)return;let i=await r.json();if(!i||i.length===0)return;t.innerHTML=`
        <h2 class="section-title">Recent Runs</h2>
        ${i.map((o,a)=>`
          <a class="run" href="/_ui${p(o.workflow_path)}?run=${p(o.id)}" style="animation-delay:${a*30}ms">
            <span class="run-dot ${o.status}"></span>
            <span class="run-name">${p(o.workflow_name)}<span class="run-id">${this.shortId(o.id)}</span></span>
            <span class="run-time">${this.relativeTime(o.started_at)}</span>
            <span class="run-status ${o.status}">${o.status}</span>
          </a>
        `).join("")}
      `}catch{}}relativeTime(e){let t=Date.now(),r=new Date(e).getTime(),i=t-r,o=Math.floor(i/1000);if(o<60)return"just now";let a=Math.floor(o/60);if(a<60)return`${a}m ago`;let n=Math.floor(a/60);if(n<24)return`${n}h ago`;return`${Math.floor(n/24)}d ago`}shortId(e){let t=e.split("_");if(t.length>=4)return t.slice(2).join("_");return e}}class be extends HTMLElement{meta;sessionId="";attachedFile=null;streamAbort=null;connectedCallback(){this.meta=JSON.parse(this.getAttribute("data-meta")||"{}");let e=new URL(window.location.href),t=e.searchParams.get("session");if(t)this.sessionId=t,this.renderChat();else this.initWithLatestSession(e)}async initWithLatestSession(e){try{let t=await fetch(`/_api/chats?workflow=${encodeURIComponent(this.meta.path)}`);if(t.ok){let r=await t.json();if(r&&r.length>0){this.sessionId=r[0].id,e.searchParams.set("session",this.sessionId),window.history.replaceState({},"",e.toString()),this.renderChat();return}}}catch{}this.sessionId=crypto.randomUUID(),e.searchParams.set("session",this.sessionId),window.history.replaceState({},"",e.toString()),this.renderChat()}renderChat(){let e=this.meta,t=this.shadowRoot||this.attachShadow({mode:"open"});t.innerHTML=`
      <style>
        ${$}
        :host { display: flex; flex-direction: row; flex: 1; overflow: hidden; position: relative; }

        /* Sidebar */
        .sidebar {
          width: 240px; flex-shrink: 0; display: flex; flex-direction: column;
          border-right: 1px solid var(--haira-border); background: var(--haira-bg);
          overflow: hidden; transition: width 0.2s, opacity 0.2s;
        }
        .sidebar.collapsed { width: 0; opacity: 0; pointer-events: none; }
        .sidebar-header {
          display: flex; align-items: center; gap: 0.4rem;
          padding: 0.55rem 0.65rem; border-bottom: 1px solid var(--haira-border); flex-shrink: 0;
        }
        .sidebar-title { font-size: 0.75rem; font-weight: 600; color: var(--haira-text-dim); flex: 1; }
        .sidebar-btn {
          background: none; border: none; color: var(--haira-muted); cursor: pointer;
          display: flex; align-items: center; justify-content: center; padding: 0.25rem;
          border-radius: 4px; transition: all 0.15s;
        }
        .sidebar-btn:hover { color: var(--haira-accent); background: var(--haira-accent-dim); }
        .sidebar-list {
          flex: 1; overflow-y: auto; padding: 0.35rem;
          display: flex; flex-direction: column; gap: 1px; ${R}
        }
        .session-item {
          display: flex; align-items: center; gap: 0.4rem; padding: 0.45rem 0.5rem;
          border-radius: 6px; cursor: pointer; transition: all 0.12s; text-decoration: none;
          color: var(--haira-text-dim); font-size: 0.78rem; line-height: 1.35; position: relative;
        }
        .session-item:hover { background: var(--haira-bg-card); color: var(--haira-text); }
        .session-item.active { background: var(--haira-accent-dim); color: var(--haira-accent); }
        .session-icon { display: flex; flex-shrink: 0; opacity: 0.5; }
        .session-item.active .session-icon { opacity: 1; }
        .session-title { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; min-width: 0; }
        .session-delete {
          display: none; background: none; border: none; color: var(--haira-muted); cursor: pointer;
          padding: 0.15rem; border-radius: 3px; flex-shrink: 0; align-items: center; justify-content: center;
        }
        .session-item:hover .session-delete { display: flex; }
        .session-delete:hover { color: var(--haira-error); background: rgba(239, 68, 68, 0.1); }
        .sidebar-empty { padding: 1rem; text-align: center; font-size: 0.75rem; color: var(--haira-muted); opacity: 0.5; }

        .sidebar-toggle {
          position: absolute; top: 0.5rem; left: 0.5rem; z-index: 10;
          background: var(--haira-bg-card); border: 1px solid var(--haira-border);
          color: var(--haira-muted); cursor: pointer; display: none;
          align-items: center; justify-content: center; padding: 0.35rem;
          border-radius: 6px; transition: all 0.15s;
        }
        .sidebar-toggle.visible { display: flex; }
        .sidebar-toggle:hover { color: var(--haira-accent); border-color: var(--haira-accent); background: var(--haira-accent-dim); }

        /* Chat main */
        .chat-main {
          flex: 1; display: flex; flex-direction: column; overflow: hidden;
          position: relative; min-width: 0; height: 100%;
        }

        /* Welcome */
        .welcome {
          flex: 1; display: flex; flex-direction: column; align-items: center;
          justify-content: center; gap: 1rem; padding: 2rem; opacity: 1; transition: opacity 0.3s;
        }
        .welcome.hidden { display: none; }
        .welcome-icon { opacity: 0.15; }
        .welcome-icon img { width: 56px; height: 56px; object-fit: contain; opacity: 1; }
        .welcome h2 { font-size: 1.1rem; font-weight: 600; color: var(--haira-text); }
        .welcome p { font-size: 0.85rem; color: var(--haira-muted); text-align: center; max-width: 420px; line-height: 1.5; }
        .suggestions { display: flex; flex-wrap: wrap; gap: 0.5rem; justify-content: center; margin-top: 0.5rem; max-width: 540px; }
        .suggestion {
          background: var(--haira-bg-card); border: 1px solid var(--haira-border);
          color: var(--haira-text-dim); padding: 0.45rem 0.85rem; border-radius: 20px;
          font-size: 0.78rem; font-family: var(--haira-font); cursor: pointer; transition: all 0.15s;
        }
        .suggestion:hover { border-color: var(--haira-accent); color: var(--haira-accent); background: var(--haira-accent-dim); }

        /* Messages */
        .messages { flex: 1; min-height: 0; overflow-y: auto; display: none; flex-direction: column; ${R} }
        .messages.active { display: flex; }
        .messages-inner {
          max-width: 768px; width: 100%; margin: 0 auto; padding: 1.5rem 1.25rem;
          display: flex; flex-direction: column; gap: 0.75rem;
        }

        /* Typing */
        .typing {
          display: none; padding: 0.25rem 0; align-items: center; gap: 0.4rem;
          font-size: 0.75rem; color: var(--haira-muted); margin-left: 2.25rem;
        }
        .typing.visible { display: flex; }
        .typing-dots { display: flex; gap: 0.2rem; align-items: center; }
        .typing-dot {
          display: inline-block; width: 5px; height: 5px; border-radius: 50%;
          background: var(--haira-accent); animation: bounce 1.4s ease-in-out infinite;
        }
        .typing-dot:nth-child(2) { animation-delay: 0.2s; }
        .typing-dot:nth-child(3) { animation-delay: 0.4s; }

        /* Drop overlay */
        .drop-overlay {
          display: none; position: absolute; inset: 0;
          background: rgba(9, 9, 11, 0.85); z-index: 200;
          align-items: center; justify-content: center; flex-direction: column; gap: 0.75rem;
          border: 2px dashed var(--haira-accent); border-radius: var(--haira-radius); margin: 0.5rem;
        }
        .drop-overlay.visible { display: flex; }
        .drop-overlay-icon { color: var(--haira-accent); opacity: 0.7; }
        .drop-overlay-text { color: var(--haira-accent); font-size: 0.9rem; font-weight: 600; }

        /* Input */
        .input-area {
          padding: 0.75rem 1rem 1rem; flex-shrink: 0;
          background: var(--haira-bg); border-top: 1px solid var(--haira-border);
        }
        .input-card {
          display: flex; flex-direction: column; background: var(--haira-bg-card);
          border: 1px solid var(--haira-border); border-radius: var(--haira-radius);
          transition: border-color 0.15s; max-width: 768px; margin: 0 auto;
        }
        .input-card:focus-within { border-color: var(--haira-border-focus); }
        .file-chip { display: none; align-items: center; gap: 0.4rem; padding: 0.4rem 0.6rem 0; margin: 0 0.5rem; }
        .file-chip.visible { display: flex; }
        .file-chip-inner {
          display: flex; align-items: center; gap: 0.35rem;
          background: var(--haira-bg-elevated); border: 1px solid var(--haira-border);
          border-radius: 6px; padding: 0.25rem 0.5rem; font-size: 0.75rem; color: var(--haira-text-dim);
        }
        .file-chip-icon { color: var(--haira-accent); display: flex; }
        .file-chip-name { max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .file-chip-size { color: var(--haira-muted); font-size: 0.7rem; }
        .file-chip-remove {
          background: none; border: none; color: var(--haira-muted); cursor: pointer;
          display: flex; padding: 0.1rem; border-radius: 3px; transition: all 0.15s;
        }
        .file-chip-remove:hover { color: var(--haira-error); background: rgba(239, 68, 68, 0.1); }
        .input-row { display: flex; align-items: flex-end; gap: 0.35rem; padding: 0.35rem; }
        .attach-btn {
          background: none; border: none; color: var(--haira-muted); cursor: pointer;
          display: flex; align-items: center; justify-content: center; padding: 0.45rem;
          border-radius: 6px; transition: all 0.15s; flex-shrink: 0;
        }
        .attach-btn:hover { color: var(--haira-accent); background: var(--haira-accent-dim); }
        textarea {
          flex: 1; background: transparent; border: none; color: var(--haira-text);
          padding: 0.5rem 0.35rem; font-size: 0.9rem; font-family: var(--haira-font);
          resize: none; min-height: 44px; max-height: 200px; outline: none; line-height: 1.5;
        }
        textarea::placeholder { color: var(--haira-muted); }
        .send-btn {
          background: var(--haira-accent); color: #1a0e04; border: none;
          width: 34px; height: 34px; border-radius: 8px; cursor: pointer;
          display: flex; align-items: center; justify-content: center;
          transition: all 0.15s; flex-shrink: 0;
        }
        .send-btn:hover { background: var(--haira-accent-light); box-shadow: 0 2px 12px rgba(232, 163, 23, 0.25); }
        .send-btn:disabled { opacity: 0.35; cursor: not-allowed; box-shadow: none; }
        .input-hint { text-align: center; font-size: 0.68rem; color: var(--haira-muted); opacity: 0.5; padding-top: 0.35rem; }

        /* Activity panel */
        .activity-panel {
          position: absolute; right: 0; top: 0; bottom: 0; width: 280px; z-index: 50;
          display: flex; flex-direction: column; border-left: 1px solid var(--haira-border);
          background: var(--haira-bg); overflow: hidden; box-shadow: -4px 0 24px rgba(0, 0, 0, 0.25);
        }
        .activity-panel.collapsed { display: none; }
        .panel-header {
          display: flex; align-items: center; gap: 0.5rem;
          padding: 0.55rem 0.75rem; border-bottom: 1px solid var(--haira-border); flex-shrink: 0;
        }
        .panel-header-icon { display: flex; color: var(--haira-muted); }
        .panel-title { font-size: 0.78rem; font-weight: 600; color: var(--haira-text-dim); flex: 1; }
        .panel-count { font-size: 0.68rem; color: var(--haira-muted); font-family: var(--haira-mono); }
        .panel-close {
          background: none; border: none; color: var(--haira-muted); cursor: pointer;
          display: flex; align-items: center; justify-content: center; padding: 0.2rem;
          border-radius: 4px; transition: all 0.15s;
        }
        .panel-close:hover { color: var(--haira-text); background: var(--haira-bg-elevated); }
        .panel-body {
          flex: 1; overflow-y: auto; padding: 0.5rem;
          display: flex; flex-direction: column; gap: 0.4rem; ${R}
        }
        .panel-empty { display: flex; align-items: center; justify-content: center; flex: 1; font-size: 0.75rem; color: var(--haira-muted); opacity: 0.5; }
        .activity-toggle {
          position: absolute; top: 0.5rem; right: 0.5rem; z-index: 10;
          background: var(--haira-bg-card); border: 1px solid var(--haira-border);
          color: var(--haira-muted); cursor: pointer; display: flex;
          align-items: center; justify-content: center; padding: 0.35rem;
          border-radius: 6px; transition: all 0.15s; gap: 0.3rem;
        }
        .activity-toggle:hover { color: var(--haira-accent); border-color: var(--haira-accent); background: var(--haira-accent-dim); }
        .activity-toggle .badge {
          display: none; min-width: 16px; height: 16px; padding: 0 4px;
          border-radius: 8px; background: var(--haira-accent); color: #1a0e04;
          font-size: 0.62rem; font-weight: 700; line-height: 16px; text-align: center;
        }
        .activity-toggle .badge.visible { display: inline-block; }

        @media (max-width: 640px) {
          .sidebar {
            position: absolute; left: 0; top: 0; bottom: 0; z-index: 100;
            box-shadow: 4px 0 24px rgba(0, 0, 0, 0.3);
          }
          .messages-inner { padding: 1rem 0.75rem; }
          .input-area { padding: 0.5rem 0.5rem 0.75rem; padding-bottom: max(0.75rem, env(safe-area-inset-bottom)); }
          .welcome { padding: 1.5rem 1rem; }
          .suggestions { max-width: 100%; }
        }
      </style>

      <div class="sidebar" id="sidebar">
        <div class="sidebar-header">
          <span class="sidebar-title">Chats</span>
          <button class="sidebar-btn" id="new-chat-btn" title="New chat">${h.plus}</button>
          <button class="sidebar-btn" id="sidebar-close-btn" title="Close sidebar">${h.chevronLeft}</button>
        </div>
        <div class="sidebar-list" id="sidebar-list">
          <div class="sidebar-empty" id="sidebar-empty">No chats yet</div>
        </div>
      </div>

      <div class="chat-main">
        <button class="sidebar-toggle" id="sidebar-open-btn" title="Show chats">${h.sidebar}</button>

        <div class="welcome" id="welcome">
          <span class="welcome-icon">${e.logo?`<img src="${p(e.logo)}" alt="">`:q.replace(/width="22" height="22"/,'width="56" height="56"')}</span>
          <h2>${p(e.title||e.name||"Chat")}</h2>
          ${e.description?`<p>${p(e.description)}</p>`:""}
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
          <span class="drop-overlay-icon">${h.attach}</span>
          <span class="drop-overlay-text">Drop file to attach</span>
        </div>

        <button class="activity-toggle" id="activity-toggle" title="Toggle activity panel">
          ${h.activity}
          <span class="badge" id="toggle-badge">0</span>
        </button>

        <div class="input-area">
          <div class="input-card" id="input-card">
            <div class="file-chip" id="file-chip">
              <div class="file-chip-inner">
                <span class="file-chip-icon">${h.file}</span>
                <span class="file-chip-name" id="file-name"></span>
                <span class="file-chip-size" id="file-size"></span>
                <button class="file-chip-remove" id="file-remove" title="Remove file">${h.xSmall}</button>
              </div>
            </div>
            <div class="input-row">
              ${e.hasFile?`<button class="attach-btn" id="attach-btn" title="Attach file">${h.attach}</button>`:""}
              <textarea id="chat-input" placeholder="${e.hasFile?"Type a message or drop a file...":"Type a message..."}" rows="1"></textarea>
              <button class="send-btn" id="send-btn" title="Send">${h.send}</button>
            </div>
          </div>
          ${e.hasFile?'<input type="file" id="file-input" style="display:none" />':""}
          <div class="input-hint">Enter to send, Shift+Enter for new line</div>
        </div>
      </div>

      <div class="activity-panel collapsed" id="activity-panel">
        <div class="panel-header">
          <span class="panel-header-icon">${h.activity}</span>
          <span class="panel-title">Activity</span>
          <span class="panel-count" id="panel-count"></span>
          <button class="panel-close" id="panel-close" title="Close panel">${h.xSmall}</button>
        </div>
        <div class="panel-body" id="panel-body">
          <div class="panel-empty" id="panel-empty">No activity yet</div>
        </div>
      </div>
    `;let r=t.getElementById("messages"),i=t.getElementById("messages-inner"),o=t.getElementById("welcome"),a=t.getElementById("chat-input"),n=t.getElementById("send-btn"),s=t.getElementById("file-chip"),l=t.getElementById("file-name"),m=t.getElementById("file-size"),u=t.getElementById("file-remove"),d=t.getElementById("typing"),c=t.getElementById("drop-overlay"),f=t.getElementById("suggestions"),g=e.hasFile?t.getElementById("attach-btn"):null,v=e.hasFile?t.getElementById("file-input"):null,E=t.getElementById("activity-panel"),C=t.getElementById("panel-body"),H=t.getElementById("panel-empty"),V=t.getElementById("panel-count"),J=t.getElementById("panel-close"),G=t.getElementById("activity-toggle"),A=t.getElementById("toggle-badge"),Fe=t.getElementById("sidebar"),O=t.getElementById("sidebar-list"),ze=t.getElementById("sidebar-empty"),De=t.getElementById("new-chat-btn"),qe=t.getElementById("sidebar-close-btn"),He=t.getElementById("sidebar-open-btn"),K=!1,N=0,ee=0;function te(b){K=b!==void 0?b:!K,E.classList.toggle("collapsed",!K)}function Re(){if(N>0)A.textContent=String(N),A.classList.add("visible");else A.classList.remove("visible");V.textContent=ee>0?String(ee):""}G.addEventListener("click",()=>te()),J.addEventListener("click",()=>te(!1));let U=!0,Me=(b)=>{U=b!==void 0?b:!U,Fe.classList.toggle("collapsed",!U),He.classList.toggle("visible",!U)};qe.addEventListener("click",()=>Me(!1)),He.addEventListener("click",()=>Me(!0));let z=this,re=async()=>{try{let b=await fetch(`/_api/chats?workflow=${encodeURIComponent(e.path)}`);if(!b.ok)return;let y=await b.json();if(!y||y.length===0){ze.style.display="",O.querySelectorAll(".session-item").forEach((x)=>x.remove());return}ze.style.display="none",O.querySelectorAll(".session-item").forEach((x)=>x.remove());for(let x of y){let S=document.createElement("div");S.className=`session-item${x.id===z.sessionId?" active":""}`,S.innerHTML=`
            <span class="session-icon">${h.chat}</span>
            <span class="session-title">${p(x.title||"New chat")}</span>
            <button class="session-delete" title="Delete">${h.trash}</button>
          `,S.addEventListener("click",(_)=>{if(_.target.closest(".session-delete"))return;z.switchSession(x.id)}),S.querySelector(".session-delete").addEventListener("click",async(_)=>{if(_.stopPropagation(),await fetch(`/_api/chats/${x.id}`,{method:"DELETE"}),x.id===z.sessionId)z.startNewChat();re()}),O.appendChild(S)}}catch{}};De.addEventListener("click",()=>{z.startNewChat()});let Ve=this.getSuggestions();for(let b of Ve){let y=document.createElement("button");y.className="suggestion",y.textContent=b,y.addEventListener("click",()=>{a.value=b,W()}),f.appendChild(y)}if(g&&v){g.addEventListener("click",()=>v.click()),v.addEventListener("change",()=>{if(v.files&&v.files[0])this.setFile(v.files[0],s,l,m)}),u.addEventListener("click",()=>{this.clearFile(s,v)});let b=0;t.addEventListener("dragenter",(y)=>{y.preventDefault(),b++,c.classList.add("visible")}),t.addEventListener("dragleave",(y)=>{if(y.preventDefault(),b--,b<=0)b=0,c.classList.remove("visible")}),t.addEventListener("dragover",(y)=>{y.preventDefault()}),t.addEventListener("drop",(y)=>{y.preventDefault(),b=0,c.classList.remove("visible");let x=y.dataTransfer;if(x&&x.files&&x.files[0])z.setFile(x.files[0],s,l,m)})}a.addEventListener("input",()=>{a.style.height="auto",a.style.height=`${Math.min(a.scrollHeight,200)}px`}),a.addEventListener("keydown",(b)=>{if(b.key==="Enter"&&!b.shiftKey)b.preventDefault(),W()}),n.addEventListener("click",W),t.addEventListener("haira-chat-input",(b)=>{let y=b.detail?.text;if(y&&!n.disabled)a.value=y,W()}),requestAnimationFrame(()=>a.focus());let Je=e.avatar||"H";function j(b,y,x){let S=document.createElement("haira-message");if(S.setAttribute("role",b),S.setAttribute("content",y),x)S.setAttribute("file",x);if(b==="assistant")S.setAttribute("avatar",Je);return i.insertBefore(S,d),r.scrollTop=r.scrollHeight,S}function Ke(){i.querySelectorAll("haira-message, haira-tool-card, haira-ui-renderer").forEach((y)=>y.remove())}async function W(){let b=a.value.trim();if(!b&&!z.attachedFile)return;a.value="",a.style.height="auto",z.streamAbort?.abort(),z.streamAbort=new AbortController,o.classList.add("hidden"),r.classList.add("active");let y=z.attachedFile?z.attachedFile.name:void 0;j("user",b,y),n.disabled=!0,d.classList.add("visible"),r.scrollTop=r.scrollHeight;let x=e.chatParam||"message",S,_={};if(_[x]=b,_.session_id=z.sessionId,z.attachedFile){let L=e.fileParam||"file_path";S=new FormData,S.append(L,z.attachedFile),S.append(x,b),S.append("session_id",z.sessionId)}if(v)z.clearFile(s,v);let I=new Map,B=null,F="";await X(e.path,_,{onToolStart:(L)=>{d.classList.remove("visible");let M=document.createElement("haira-tool-card");if(H.style.display="none",C.appendChild(M),M.setTool(L.tool),I.set(L.tool,M),C.scrollTop=C.scrollHeight,N++,ee++,Re(),!K)te(!0)},onToolRender:(L)=>{let M=document.createElement("haira-ui-renderer");i.insertBefore(M,d),M.render(L),r.scrollTop=r.scrollHeight},onToolEnd:(L)=>{let M=I.get(L.tool);if(M)M.complete(L.ok!==!1),I.delete(L.tool);d.classList.add("visible"),N=Math.max(0,N-1),Re()},onDelta:(L)=>{if(d.classList.remove("visible"),!B)B=j("assistant","");F+=L,B.updateContent(F),r.scrollTop=r.scrollHeight},onError:(L)=>{if(d.classList.remove("visible"),!B)B=j("assistant","");B.updateContent(`Error: ${L}`),n.disabled=!1,a.focus()},onDone:()=>{if(d.classList.remove("visible"),!B&&F==="")B=j("assistant",""),B.updateContent("No response received. Please check the server logs.");n.disabled=!1,a.focus(),re()}},S,z.streamAbort?.signal)}(async(b)=>{try{let y=await fetch(`/_api/chats/${b}`);if(!y.ok)return;let x=await y.json();if(!x.messages||x.messages.length===0)return;o.classList.add("hidden"),r.classList.add("active"),Ke();let S=x.messages;for(let _=0;_<S.length;_++){let I=S[_],B=j(I.role,I.content);if(I.role==="assistant")B.updateContent(I.content);if(I.ui_events&&I.ui_events.length>0){let F=S.slice(_+1).some((L)=>L.role==="user");for(let L of I.ui_events){let M=document.createElement("haira-ui-renderer");if(F)M.setAttribute("data-restored","true");i.insertBefore(M,d),M.render(L)}}}r.scrollTop=r.scrollHeight}catch{}})(this.sessionId),re()}switchSession(e){this.streamAbort?.abort(),this.streamAbort=null,this.sessionId=e;let t=new URL(window.location.href);if(t.searchParams.set("session",e),window.history.pushState({},"",t.toString()),this.shadowRoot)this.shadowRoot.innerHTML="";this.renderChat()}startNewChat(){let e=crypto.randomUUID();this.switchSession(e)}setFile(e,t,r,i){this.attachedFile=e,r.textContent=e.name,i.textContent=ne(e.size),t.classList.add("visible")}clearFile(e,t){this.attachedFile=null,e.classList.remove("visible"),t.value=""}getSuggestions(){if(this.meta.suggestions&&this.meta.suggestions.length>0)return this.meta.suggestions;return["What can you help me with?","Hello!"]}}class ge extends w{startTime=0;render(){return`
      <div class="card" id="card">
        <div class="icon running" id="icon">${h.spinner}</div>
        <div class="info">
          <div class="tool-name" id="name"></div>
          <div class="tool-status" id="status">Running...</div>
        </div>
        <span class="duration" id="duration"></span>
      </div>`}styles(){return`
      ${k}
      .card {
        display: flex; align-items: center; gap: 0.5rem;
        padding: 0.5rem 0.75rem;
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: 8px;
      }
      .icon {
        display: flex; align-items: center; justify-content: center;
        width: 24px; height: 24px; border-radius: 6px; flex-shrink: 0;
      }
      .icon.running { background: rgba(232, 163, 23, 0.1); color: var(--haira-accent); }
      .icon.done { background: rgba(34, 197, 94, 0.1); color: var(--haira-success); }
      .icon.failed { background: rgba(239, 68, 68, 0.1); color: var(--haira-error); }
      .info { flex: 1; min-width: 0; }
      .tool-name { font-size: 0.78rem; font-weight: 600; color: var(--haira-text); }
      .tool-status { font-size: 0.7rem; color: var(--haira-muted); display: flex; align-items: center; gap: 0.3rem; }
      .duration { font-family: var(--haira-mono); font-size: 0.68rem; color: var(--haira-muted); flex-shrink: 0; }
      @keyframes spin { to { transform: rotate(360deg); } }`}onMount(){this.startTime=Date.now()}setTool(e){let t=this.root?.getElementById("name");if(t)t.textContent=se(e)}complete(e){let t=this.root?.getElementById("icon"),r=this.root?.getElementById("status"),i=this.root?.getElementById("duration"),a=((Date.now()-this.startTime)/1000).toFixed(1);if(t)t.className=`icon ${e?"done":"failed"}`,t.innerHTML=e?h.check:h.x;if(r)r.textContent=e?"Completed":"Failed";if(i)i.textContent=`${a}s`}}var Z={success:h.statusSuccess,error:h.statusError,warning:h.statusWarning,info:h.statusInfo},Ae={success:"var(--haira-success)",error:"var(--haira-error)",warning:"var(--haira-warn)",info:"var(--haira-info)"};class ve extends w{render(){return`
      <div class="card" id="card">
        <div class="header" id="header">
          <span class="icon" id="icon"></span>
          <span class="title" id="title"></span>
        </div>
        <div class="message" id="message"></div>
        <div class="sections" id="sections" style="display:none"></div>
      </div>`}styles(){return`
      ${k}
      .card { ${T} }
      .header {
        display: flex; align-items: center; gap: 0.4rem;
        padding: 0.45rem 0.75rem;
      }
      .icon { display: flex; align-items: center; flex-shrink: 0; }
      .icon svg { width: 14px; height: 14px; }
      .title { font-size: 0.78rem; font-weight: 600; }
      .message {
        font-size: 0.75rem; color: var(--haira-text-dim);
        padding: 0 0.75rem 0.45rem 2rem; line-height: 1.4;
      }
      .sections { border-top: 1px solid var(--haira-border); }
      .section {
        padding: 0.4rem 0.75rem;
        border-bottom: 1px solid var(--haira-border);
      }
      .section:last-child { border-bottom: none; }
      .section-label {
        font-size: 0.68rem; font-weight: 600; color: var(--haira-muted);
        text-transform: uppercase; letter-spacing: 0.04em; margin-bottom: 0.2rem;
      }
      .section-content {
        font-size: 0.75rem; color: var(--haira-text-dim);
        line-height: 1.4; white-space: pre-wrap;
      }
      .section-content.code {
        font-family: var(--haira-mono); font-size: 0.72rem;
        background: var(--haira-bg); padding: 0.35rem 0.6rem;
        border-radius: var(--haira-radius-sm); overflow-x: auto;
      }
      .card.inline .header { padding: 0.35rem 0.65rem; }
      .card.inline .message { display: inline; padding: 0; margin-left: 0.15rem; font-weight: 400; }
      .card.inline .header-row { display: flex; align-items: center; gap: 0.4rem; flex-wrap: wrap; }`}onUpdate(){let{status:e="info",title:t="",message:r,sections:i}=this.props,o=Ae[e]||Ae.info,a=i&&i.length>0,n=!!r,s=this.$("card"),l=this.$("header");if(!a&&n){s.classList.add("inline"),l.innerHTML=`
        <div class="header-row">
          <span class="icon" id="icon"></span>
          <span class="title" id="title"></span>
          <span class="message" id="message"></span>
        </div>`;let f=l.querySelector("#icon");f.innerHTML=Z[e]||Z.info,f.style.color=o;let g=l.querySelector("#title");g.textContent=t,g.style.color=o,l.querySelector("#message").textContent=r,this.$("message").style.display="none",this.$("sections").style.display="none",s.style.borderLeft=`3px solid ${o}`;return}s.classList.remove("inline");let m=this.$("icon");m.innerHTML=Z[e]||Z.info,m.style.color=o;let u=this.$("title");u.textContent=t,u.style.color=o;let d=this.$("message");if(n)d.textContent=r,d.style.display="";else d.style.display="none";s.style.borderLeft=`3px solid ${o}`;let c=this.$("sections");if(a)c.style.display="",c.innerHTML=i.map((f)=>`
          <div class="section">
            <div class="section-label">${this.esc(f.label||"")}</div>
            <div class="section-content ${f.style==="code"?"code":""}">${this.esc(f.content||"")}</div>
          </div>`).join("");else c.style.display="none"}}var Ne=15;class ye extends w{allRows=[];filteredRows=[];headers=[];highlight=new Set;searchTerm="";tabs=[];activeTab=0;render(){return`
      <div class="card">
        <div class="toolbar" id="toolbar" style="display:none">
          <span class="toolbar-title" id="title"></span>
          <span class="row-count" id="row-count"></span>
          <div class="search-wrap" id="search-wrap" style="display:none">
            ${h.search}
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
      </div>`}styles(){return`
      ${k}
      .card { ${T} }
      .toolbar {
        display: flex; align-items: center; gap: 0.5rem;
        padding: 0.5rem 0.75rem; border-bottom: 1px solid var(--haira-border);
      }
      .toolbar-title { font-size: 0.78rem; font-weight: 600; color: var(--haira-text); white-space: nowrap; }
      .row-count {
        font-size: 0.68rem; color: var(--haira-muted); background: var(--haira-bg);
        padding: 0.15rem 0.45rem; border-radius: 9px; white-space: nowrap; flex-shrink: 0;
      }
      .search-wrap { margin-left: auto; position: relative; flex-shrink: 0; }
      .search-wrap svg { position: absolute; left: 0.45rem; top: 50%; transform: translateY(-50%); color: var(--haira-muted); pointer-events: none; }
      .search {
        background: var(--haira-bg); border: 1px solid var(--haira-border);
        color: var(--haira-text); font-size: 0.72rem; font-family: var(--haira-font);
        padding: 0.28rem 0.5rem 0.28rem 1.6rem; border-radius: 6px; width: 160px;
        outline: none; transition: border-color 0.15s;
      }
      .search:focus { border-color: var(--haira-accent); }
      .search::placeholder { color: var(--haira-muted); }
      .tab-bar {
        display: flex; gap: 0; border-bottom: 1px solid var(--haira-border);
        background: var(--haira-bg); overflow-x: auto; scrollbar-width: none;
      }
      .tab-bar::-webkit-scrollbar { display: none; }
      .tab {
        padding: 0.4rem 0.75rem; font-size: 0.72rem; font-family: var(--haira-font);
        color: var(--haira-muted); background: none; border: none;
        border-bottom: 2px solid transparent; cursor: pointer;
        white-space: nowrap; transition: color 0.15s, border-color 0.15s; flex-shrink: 0;
      }
      .tab:hover { color: var(--haira-text); }
      .tab.active { color: var(--haira-accent); border-bottom-color: var(--haira-accent); font-weight: 600; }
      .table-scroll { overflow: auto; ${R} }
      .table-scroll.capped { max-height: 420px; }
      table { width: 100%; border-collapse: collapse; font-size: 0.74rem; }
      th {
        text-align: left; padding: 0.35rem 0.65rem; font-weight: 600;
        font-size: 0.68rem; color: var(--haira-muted); text-transform: uppercase;
        letter-spacing: 0.04em; background: var(--haira-bg);
        border-bottom: 1px solid var(--haira-border); white-space: nowrap;
        position: sticky; top: 0; z-index: 1;
      }
      td {
        padding: 0.28rem 0.65rem; color: var(--haira-text-dim);
        border-bottom: 1px solid var(--haira-border); line-height: 1.35;
        max-width: 320px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
      }
      td:hover { white-space: normal; word-break: break-all; }
      tr:last-child td { border-bottom: none; }
      tr:hover td { background: var(--haira-bg-card-hover); }
      tr.highlight td { background: rgba(232, 163, 23, 0.06); }
      .no-results { padding: 1.5rem; text-align: center; color: var(--haira-muted); font-size: 0.75rem; }
      .footer {
        padding: 0.3rem 0.75rem; border-top: 1px solid var(--haira-border);
        font-size: 0.68rem; color: var(--haira-muted); text-align: right;
      }
      @media (max-width: 640px) {
        .toolbar { flex-wrap: wrap; }
        .search-wrap { margin-left: 0; width: 100%; }
        .search { width: 100%; }
        td { max-width: 200px; }
      }`}onMount(){this.$("search").addEventListener("input",(e)=>{this.searchTerm=e.target.value.toLowerCase(),this.applyFilter()})}onUpdate(){try{let{title:e,tabs:t}=this.props;if(t&&t.length>0)this.tabs=t,this.activeTab=0,this.$("toolbar").style.display="",this.$("title").textContent=e||"Table",this.renderTabBar(),this.loadTab(0);else this.tabs=[],this.loadSingleTable()}catch{}}loadSingleTable(){let{title:e,headers:t=[],rows:r=[],highlight:i=[]}=this.props;this.headers=t,this.allRows=r,this.highlight=new Set(i);let o=this.$("toolbar"),a=this.$("title"),n=this.$("row-count"),s=this.$("search-wrap"),l=!!e,m=this.allRows.length>=Ne;if(l||m)o.style.display="",a.textContent=e||"Table",n.textContent=`${this.allRows.length} rows`;if(m)s.style.display="";this.$("scroll").classList.toggle("capped",m),this.$("thead").innerHTML=`<tr>${this.headers.map((u)=>`<th>${this.esc(u)}</th>`).join("")}</tr>`,this.searchTerm="",this.$("search").value="",this.applyFilter()}renderTabBar(){let e=this.$("tab-bar");e.style.display="flex",e.innerHTML="";for(let t=0;t<this.tabs.length;t++){let r=this.tabs[t],i=document.createElement("button");i.className=`tab${t===this.activeTab?" active":""}`,i.textContent=`${r.name} (${r.rows.length})`,i.addEventListener("click",()=>this.loadTab(t)),e.appendChild(i)}}loadTab(e){this.activeTab=e;let t=this.tabs[e];this.headers=t.headers||[],this.allRows=t.rows||[],this.highlight=new Set(t.highlight||[]),this.$("tab-bar").querySelectorAll(".tab").forEach((i,o)=>{i.classList.toggle("active",o===e)}),this.$("row-count").textContent=`${this.allRows.length} rows`;let r=this.allRows.length>=Ne;this.$("search-wrap").style.display=r?"":"none",this.$("scroll").classList.toggle("capped",r),this.$("thead").innerHTML=`<tr>${this.headers.map((i)=>`<th>${this.esc(i)}</th>`).join("")}</tr>`,this.searchTerm="",this.$("search").value="",this.applyFilter()}applyFilter(){if(this.searchTerm)this.filteredRows=this.allRows.map((e,t)=>({row:e,idx:t})).filter(({row:e})=>e.some((t)=>String(t).toLowerCase().includes(this.searchTerm)));else this.filteredRows=this.allRows.map((e,t)=>({row:e,idx:t}));this.renderRows()}renderRows(){let e=this.$("tbody"),t=this.$("footer"),r=this.$("row-count");if(this.filteredRows.length===0&&this.searchTerm)e.innerHTML=`<tr><td colspan="${this.headers.length||1}" class="no-results">No matching rows</td></tr>`;else e.innerHTML=this.filteredRows.map(({row:i,idx:o})=>`<tr class="${this.highlight.has(o)?"highlight":""}">${i.map((a)=>`<td title="${this.esc(String(a))}">${this.esc(String(a))}</td>`).join("")}</tr>`).join("");if(this.searchTerm)r.textContent=`${this.filteredRows.length} / ${this.allRows.length} rows`,t.style.display="",t.textContent=`Showing ${this.filteredRows.length} of ${this.allRows.length} rows`;else r.textContent=`${this.allRows.length} rows`,t.style.display="none"}}class xe extends w{codeText="";tabs=[];activeTab=0;render(){return`
      <div class="card">
        <div class="header">
          <div style="display:flex;align-items:center;gap:0.5rem">
            <span class="title" id="title"></span>
            <span class="lang" id="lang"></span>
          </div>
          <div class="actions">
            <button class="copy-btn" id="copy-btn">${h.copy} Copy</button>
          </div>
        </div>
        <div class="tab-bar" id="tab-bar" style="display:none"></div>
        <div class="code-scroll" id="code-scroll">
          <pre><code id="code"></code></pre>
        </div>
      </div>`}styles(){return`
      ${k}
      .card { ${T} }
      .header {
        display: flex; align-items: center; justify-content: space-between;
        padding: 0.45rem 0.75rem; border-bottom: 1px solid var(--haira-border);
        background: var(--haira-bg);
      }
      .title { font-size: 0.78rem; font-weight: 600; color: var(--haira-text); }
      .lang { font-size: 0.68rem; color: var(--haira-muted); font-family: var(--haira-mono); }
      .actions { display: flex; align-items: center; gap: 0.5rem; }
      .copy-btn {
        background: none; border: none; color: var(--haira-muted);
        cursor: pointer; display: flex; align-items: center; gap: 0.3rem;
        font-size: 0.7rem; font-family: var(--haira-font);
        padding: 0.2rem 0.4rem; border-radius: 4px; transition: all 0.15s;
      }
      .copy-btn:hover { color: var(--haira-accent); background: var(--haira-accent-dim); }
      .copy-btn.copied { color: var(--haira-success); }
      .tab-bar {
        display: flex; gap: 0; border-bottom: 1px solid var(--haira-border);
        background: var(--haira-bg); overflow-x: auto; scrollbar-width: none;
      }
      .tab-bar::-webkit-scrollbar { display: none; }
      .tab {
        padding: 0.4rem 0.75rem; font-size: 0.72rem; font-family: var(--haira-font);
        color: var(--haira-muted); background: none; border: none;
        border-bottom: 2px solid transparent; cursor: pointer;
        white-space: nowrap; transition: color 0.15s, border-color 0.15s; flex-shrink: 0;
      }
      .tab:hover { color: var(--haira-text); }
      .tab.active { color: var(--haira-accent); border-bottom-color: var(--haira-accent); font-weight: 600; }
      .code-scroll { max-height: 480px; overflow: auto; ${R} }
      pre { margin: 0; padding: 0.75rem 1rem; }
      code {
        font-family: var(--haira-mono); font-size: 0.78rem;
        color: var(--haira-text-dim); line-height: 1.6; white-space: pre;
      }`}onMount(){this.$("copy-btn").addEventListener("click",()=>this.copyCode())}onUpdate(){let{title:e,language:t,code:r,tabs:i}=this.props;if(this.$("title").textContent=e||"",i&&i.length>0)this.tabs=i,this.activeTab=0,this.renderTabBar(),this.loadTab(0);else this.tabs=[],this.$("lang").textContent=t||"",this.codeText=r||"",this.$("code").textContent=this.codeText}renderTabBar(){let e=this.$("tab-bar");e.style.display="flex",e.innerHTML="";for(let t=0;t<this.tabs.length;t++){let r=document.createElement("button");r.className=`tab${t===this.activeTab?" active":""}`,r.textContent=this.tabs[t].name,r.addEventListener("click",()=>this.loadTab(t)),e.appendChild(r)}}loadTab(e){this.activeTab=e;let t=this.tabs[e];this.$("tab-bar").querySelectorAll(".tab").forEach((r,i)=>{r.classList.toggle("active",i===e)}),this.$("lang").textContent=t.language||"",this.codeText=t.code||"",this.$("code").textContent=this.codeText,this.$("code-scroll").scrollTop=0}copyCode(){navigator.clipboard.writeText(this.codeText).then(()=>{let e=this.$("copy-btn");e.innerHTML=`${h.copyDone} Copied`,e.classList.add("copied"),setTimeout(()=>{e.innerHTML=`${h.copy} Copy`,e.classList.remove("copied")},2000)})}}class we extends w{render(){return`
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
      </div>`}styles(){return`
      ${k}
      .card { ${T} }
      .title-bar {
        padding: 0.6rem 1rem;
        font-size: 0.8rem;
        font-weight: 600;
        color: var(--haira-text);
        border-bottom: 1px solid var(--haira-border);
        display: none;
      }
      .diff-grid { display: grid; grid-template-columns: 1fr 1fr; }
      .pane { overflow-x: auto; ${R} }
      .pane + .pane { border-left: 1px solid var(--haira-border); }
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
      .pane:first-child pre { background: rgba(239, 68, 68, 0.03); }
      .pane:last-child pre { background: rgba(34, 197, 94, 0.03); }`}onUpdate(){let e=this.props,t=this.$("title");if(e.title)t.textContent=e.title,t.style.display="";this.$("before-label").textContent=e.before_label||"Before",this.$("after-label").textContent=e.after_label||"After",this.$("before").textContent=e.before||"",this.$("after").textContent=e.after||""}}var Ue={success:"var(--haira-success)",error:"var(--haira-error)",warning:"var(--haira-warn)",info:"var(--haira-info)",code:"inherit"};class $e extends w{render(){return`
      <div class="card">
        <div class="title-bar" id="title"></div>
        <div class="items" id="items"></div>
      </div>`}styles(){return`
      ${k}
      .card { ${T} }
      .title-bar {
        padding: 0.6rem 1rem;
        font-size: 0.8rem;
        font-weight: 600;
        color: var(--haira-text);
        border-bottom: 1px solid var(--haira-border);
        display: none;
      }
      .items { padding: 0.5rem 0; }
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
      }`}onUpdate(){let{title:e,items:t=[]}=this.props,r=this.$("title");if(e)r.textContent=e,r.style.display="";this.$("items").innerHTML=t.map((i)=>{let o=Ue[i.style||""]||"",a=o&&o!=="inherit"?`color:${o}`:"",n=i.style==="code";return`<div class="item">
          <span class="key">${this.esc(i.key||"")}</span>
          <span class="value ${n?"code":""}" style="${a}">${this.esc(i.value||"")}</span>
        </div>`}).join("")}}var je={done:{icon:h.stepDone,color:"var(--haira-success)"},active:{icon:h.stepActive,color:"var(--haira-accent)"},pending:{icon:h.stepPending,color:"var(--haira-muted)"},failed:{icon:h.stepFailed,color:"var(--haira-error)"}};class ke extends w{render(){return`
      <div class="card">
        <div class="title-bar" id="title"></div>
        <div class="steps" id="steps"></div>
      </div>`}styles(){return`
      ${k}
      .card { ${T} }
      .title-bar {
        padding: 0.6rem 1rem;
        font-size: 0.8rem;
        font-weight: 600;
        color: var(--haira-text);
        border-bottom: 1px solid var(--haira-border);
        display: none;
      }
      .steps { padding: 0.6rem 1rem; }
      .step {
        display: flex; align-items: flex-start; gap: 0.6rem;
        position: relative; padding-bottom: 0.75rem;
      }
      .step:last-child { padding-bottom: 0; }
      .step::before {
        content: ""; position: absolute;
        left: 6.5px; top: 18px; bottom: 0;
        width: 1px; background: var(--haira-border);
      }
      .step:last-child::before { display: none; }
      .step-icon { display: flex; flex-shrink: 0; margin-top: 1px; }
      .step-content { flex: 1; min-width: 0; }
      .step-name { font-size: 0.8rem; font-weight: 500; line-height: 1.3; }
      .step-detail { font-size: 0.72rem; color: var(--haira-muted); margin-top: 0.15rem; }
      @keyframes spin { to { transform: rotate(360deg); } }`}onUpdate(){let{title:e,steps:t=[]}=this.props,r=this.$("title");if(e)r.textContent=e,r.style.display="";this.$("steps").innerHTML=t.map((i)=>{let o=je[i.status]||je.pending;return`<div class="step">
          <span class="step-icon" style="color:${o.color}">${o.icon}</span>
          <div class="step-content">
            <div class="step-name" style="color:${o.color}">${this.esc(i.name||"")}</div>
            ${i.detail?`<div class="step-detail">${this.esc(i.detail)}</div>`:""}
          </div>
        </div>`}).join("")}}class Se extends w{render(){return`
      <div class="card">
        <div class="title-bar" id="title"></div>
        <div class="fields" id="fields"></div>
        <div class="submit-area">
          <button class="submit-btn" id="submit-btn">Submit</button>
        </div>
      </div>`}styles(){return`
      ${k}
      .card { ${T} }
      .title-bar {
        padding: 0.6rem 1rem; font-size: 0.8rem; font-weight: 600;
        color: var(--haira-text); border-bottom: 1px solid var(--haira-border); display: none;
      }
      .fields { padding: 0.75rem 1rem; display: flex; flex-direction: column; gap: 0.6rem; }
      .field-label { font-size: 0.75rem; font-weight: 600; color: var(--haira-text-dim); margin-bottom: 0.2rem; }
      .field-label .required { color: var(--haira-error); margin-left: 0.2rem; }
      input, select, textarea {
        width: 100%; background: var(--haira-bg); border: 1px solid var(--haira-border);
        color: var(--haira-text); padding: 0.45rem 0.65rem;
        border-radius: var(--haira-radius-sm); font-size: 0.8rem;
        font-family: var(--haira-font); outline: none; transition: border-color 0.15s;
      }
      input:focus, select:focus, textarea:focus { border-color: var(--haira-border-focus); }
      textarea { min-height: 60px; resize: vertical; }
      select {
        cursor: pointer; appearance: none;
        background-image: url("data:image/svg+xml,%3Csvg width='10' height='6' viewBox='0 0 10 6' fill='none' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M1 1l4 4 4-4' stroke='%2371717a' stroke-width='1.5' stroke-linecap='round'/%3E%3C/svg%3E");
        background-repeat: no-repeat; background-position: right 0.6rem center; padding-right: 2rem;
      }
      .submit-area { padding: 0.5rem 1rem 0.75rem; border-top: 1px solid var(--haira-border); }
      .submit-btn {
        background: var(--haira-accent); color: #1a0e04; border: none;
        padding: 0.5rem 1.2rem; border-radius: var(--haira-radius-sm);
        font-size: 0.8rem; font-weight: 600; font-family: var(--haira-font);
        cursor: pointer; transition: all 0.15s;
      }
      .submit-btn:hover { background: var(--haira-accent-light); box-shadow: 0 2px 12px rgba(232, 163, 23, 0.25); }`}onUpdate(){let{title:e,fields:t=[],submit_label:r="Submit",submit_action:i=""}=this.props,o=this.$("title");if(e)o.textContent=e,o.style.display="";this.$("submit-btn").textContent=r;let a=this.$("fields");a.innerHTML=t.map((n)=>{let s=n.name||"",l=n.label||s,m=n.field_type||"text",u=n.value||"",d=n.required,c=n.options||[],f;if(m==="select"&&c.length>0)f=`<select name="${this.escAttr(s)}">
            ${c.map((g)=>`<option value="${this.escAttr(g)}" ${g===u?"selected":""}>${this.esc(g)}</option>`).join("")}
          </select>`;else if(m==="textarea")f=`<textarea name="${this.escAttr(s)}">${this.esc(u)}</textarea>`;else f=`<input type="${this.escAttr(m)}" name="${this.escAttr(s)}" value="${this.escAttr(u)}" ${d?"required":""} />`;return`<div class="field-group">
          <div class="field-label">${this.esc(l)}${d?'<span class="required">*</span>':""}</div>
          ${f}
        </div>`}).join(""),this.$("submit-btn").onclick=()=>{let n={};a.querySelectorAll("input, select, textarea").forEach((s)=>{let l=s;n[l.name]=l.value}),this.emit("haira-form-submit",{action:i,data:n})}}}class Te extends w{answered=!1;render(){return`
      <div class="card" id="card">
        <div class="title" id="title"></div>
        <div class="message" id="message"></div>
        <div class="actions" id="actions">
          <button class="confirm-btn" id="confirm-btn"></button>
          <button class="deny-btn" id="deny-btn"></button>
        </div>
        <div class="selected-indicator" id="indicator"></div>
      </div>`}styles(){return`
      ${k}
      .card {
        ${T}
        border-left: 3px solid var(--haira-info);
        padding: 0.55rem 0.75rem;
      }
      .title { font-size: 0.78rem; font-weight: 600; color: var(--haira-text); margin-bottom: 0.15rem; }
      .message { font-size: 0.73rem; color: var(--haira-text-dim); line-height: 1.4; margin-bottom: 0.5rem; }
      .actions { display: flex; gap: 0.4rem; }
      button {
        font-family: var(--haira-font); font-size: 0.73rem; font-weight: 600;
        padding: 0.32rem 0.85rem; border-radius: 6px; cursor: pointer; transition: all 0.15s;
      }
      .confirm-btn { background: var(--haira-accent); color: #1a0e04; border: none; }
      .confirm-btn:hover { background: var(--haira-accent-light); }
      .deny-btn { background: transparent; color: var(--haira-text-dim); border: 1px solid var(--haira-border); }
      .deny-btn:hover { border-color: var(--haira-text-dim); color: var(--haira-text); }
      button:disabled { opacity: 0.4; cursor: default; pointer-events: none; }
      button.selected { opacity: 1; }
      .selected-indicator { display: none; font-size: 0.68rem; color: var(--haira-muted); margin-top: 0.35rem; }
      .selected-indicator.visible { display: block; }`}onUpdate(){let{title:e="Confirm",message:t,confirm_label:r="Confirm",deny_label:i="Cancel",_restored:o}=this.props;this.$("title").textContent=e;let a=this.$("message");if(t)a.textContent=t,a.style.display="";else a.style.display="none";let n=this.$("confirm-btn"),s=this.$("deny-btn");if(n.textContent=r,s.textContent=i,o)this.answered=!0,n.disabled=!0,s.disabled=!0;else n.onclick=()=>this.select(r,n,s),s.onclick=()=>this.select(i,s,n)}select(e,t,r){if(this.answered)return;this.answered=!0,t.classList.add("selected"),t.disabled=!0,r.disabled=!0;let i=this.$("indicator");i.textContent=`Selected: ${e}`,i.classList.add("visible");let o=this.$("title")?.textContent||"";this.emit("haira-chat-input",{text:`[User clicked "${e}" on confirmation: ${o}]`})}}class Ce extends w{answered=!1;render(){return`
      <div class="card">
        <div class="title" id="title"></div>
        <div id="options"></div>
      </div>`}styles(){return`
      ${k}
      .card { ${T} padding: 0.55rem 0.75rem; }
      .title { font-size: 0.78rem; font-weight: 600; color: var(--haira-text); margin-bottom: 0.45rem; }
      .options-buttons { display: flex; flex-wrap: wrap; gap: 0.35rem; }
      .opt-btn {
        background: transparent; border: 1px solid var(--haira-border);
        color: var(--haira-text-dim); font-family: var(--haira-font);
        font-size: 0.73rem; padding: 0.3rem 0.7rem; border-radius: 16px;
        cursor: pointer; transition: all 0.15s;
      }
      .opt-btn:hover { border-color: var(--haira-accent); color: var(--haira-accent); background: var(--haira-accent-dim); }
      .opt-btn:disabled { opacity: 0.35; cursor: default; pointer-events: none; }
      .opt-btn.selected { opacity: 1; background: var(--haira-accent); color: #1a0e04; border-color: var(--haira-accent); }
      .options-list { display: flex; flex-direction: column; gap: 0.15rem; }
      .opt-row {
        display: flex; align-items: center; gap: 0.45rem;
        padding: 0.35rem 0.5rem; border-radius: 6px; cursor: pointer;
        transition: background 0.15s; font-size: 0.75rem; color: var(--haira-text-dim);
      }
      .opt-row:hover { background: var(--haira-bg-card-hover); }
      .opt-radio {
        width: 14px; height: 14px; border-radius: 50%;
        border: 2px solid var(--haira-border); flex-shrink: 0;
        transition: all 0.15s; display: flex; align-items: center; justify-content: center;
      }
      .opt-row:hover .opt-radio { border-color: var(--haira-accent); }
      .opt-row.selected .opt-radio { border-color: var(--haira-accent); background: var(--haira-accent); }
      .opt-row.selected .opt-radio::after { content: ""; width: 5px; height: 5px; border-radius: 50%; background: #1a0e04; }
      .opt-row.disabled { opacity: 0.35; cursor: default; pointer-events: none; }
      .opt-row.selected.disabled { opacity: 1; }`}onUpdate(){let{title:e="Choose an option",options:t=[],style:r="buttons",_restored:i}=this.props;this.$("title").textContent=e;let o=this.$("options");if(i)this.answered=!0;if(r==="list"){if(o.className="options-list",o.innerHTML=t.map((a)=>`<div class="opt-row${i?" disabled":""}" data-value="${this.escAttr(a)}">
          <span class="opt-radio"></span><span>${this.esc(a)}</span>
        </div>`).join(""),!i)o.querySelectorAll(".opt-row").forEach((a)=>{a.addEventListener("click",()=>this.selectOption(a.dataset.value||"",o,"list"))})}else if(o.className="options-buttons",o.innerHTML=t.map((a)=>`<button class="opt-btn" data-value="${this.escAttr(a)}"${i?" disabled":""}>${this.esc(a)}</button>`).join(""),!i)o.querySelectorAll(".opt-btn").forEach((a)=>{a.addEventListener("click",()=>this.selectOption(a.dataset.value||"",o,"buttons"))})}selectOption(e,t,r){if(this.answered)return;if(this.answered=!0,r==="list")t.querySelectorAll(".opt-row").forEach((o)=>{let a=o;if(a.dataset.value===e)a.classList.add("selected","disabled");else a.classList.add("disabled")});else t.querySelectorAll(".opt-btn").forEach((o)=>{let a=o;if(a.dataset.value===e)a.classList.add("selected");a.disabled=!0});let i=this.$("title")?.textContent||"";this.emit("haira-chat-input",{text:`[User selected "${e}" from choices: ${i}]`})}}var P=["#e8a317","#3b82f6","#22c55e","#ef4444","#a855f7","#f59e0b","#06b6d4","#ec4899","#84cc16","#f97316"];class Le extends w{canvas;ctx;render(){return`
      <div class="card">
        <div class="header" id="header" style="display:none">
          <span class="title" id="title"></span>
        </div>
        <div class="canvas-wrap">
          <canvas id="canvas"></canvas>
        </div>
        <div class="legend" id="legend"></div>
      </div>`}styles(){return`
      ${k}
      .card { ${T} }
      .header {
        padding: 0.5rem 0.75rem; border-bottom: 1px solid var(--haira-border);
        font-size: 0.78rem; font-weight: 600; color: var(--haira-text);
      }
      .canvas-wrap { padding: 0.75rem; }
      canvas { width: 100%; display: block; }
      .legend {
        display: flex; flex-wrap: wrap; gap: 0.75rem; padding: 0 0.75rem 0.6rem;
        font-size: 0.7rem; color: var(--haira-text-dim);
      }
      .legend-item { display: flex; align-items: center; gap: 0.3rem; }
      .legend-dot { width: 8px; height: 8px; border-radius: 2px; flex-shrink: 0; }`}onMount(){this.canvas=this.$("canvas"),this.ctx=this.canvas.getContext("2d")}onUpdate(){let{title:e,type:t,labels:r=[],datasets:i=[],height:o=240}=this.props,a=this.$("header");if(e)this.$("title").textContent=e,a.style.display="";let n=window.devicePixelRatio||1,s=this.canvas.parentElement.clientWidth;this.canvas.width=s*n,this.canvas.height=o*n,this.canvas.style.height=`${o}px`,this.ctx.scale(n,n);let l=i.map((u,d)=>({...u,color:u.color||P[d%P.length]}));switch(t){case"bar":this.drawBar(r,l,s,o);break;case"pie":this.drawPie(l,s,o);break;case"line":case"area":this.drawLine(r,l,s,o,t==="area");break;case"scatter":this.drawScatter(r,l,s,o);break;default:this.drawBar(r,l,s,o)}let m=this.$("legend");if(l.length>1||t==="pie"){let u=t==="pie"?r:l.map((c)=>c.label),d=t==="pie"?r.map((c,f)=>P[f%P.length]):l.map((c)=>c.color);m.innerHTML=u.map((c,f)=>`<span class="legend-item"><span class="legend-dot" style="background:${d[f]}"></span>${this.esc(c)}</span>`).join("")}else m.innerHTML=""}drawBar(e,t,r,i){let o=this.ctx,a={top:10,right:10,bottom:30,left:45},n=r-a.left-a.right,s=i-a.top-a.bottom,l=t.flatMap((v)=>v.data),m=Math.max(...l,0)*1.1||1;o.strokeStyle="rgba(63, 63, 70, 0.3)",o.lineWidth=0.5;for(let v=0;v<=4;v++){let E=a.top+s-v/4*s;o.beginPath(),o.moveTo(a.left,E),o.lineTo(r-a.right,E),o.stroke(),o.fillStyle="#71717a",o.font="10px -apple-system, sans-serif",o.textAlign="right",o.fillText(this.formatNum(m/4*v),a.left-5,E+3)}let u=e.length,d=t.length,c=n/u,f=Math.min(c*0.7/d,40),g=f*d;for(let v=0;v<u;v++){let E=a.left+v*c+(c-g)/2;for(let H=0;H<d;H++){let J=(t[H].data[v]||0)/m*s,G=E+H*f,A=a.top+s-J;o.fillStyle=t[H].color,o.beginPath(),o.roundRect(G,A,f-1,J,[3,3,0,0]),o.fill()}o.fillStyle="#71717a",o.font="10px -apple-system, sans-serif",o.textAlign="center";let C=a.left+v*c+c/2;o.fillText(this.truncLabel(e[v],10),C,i-8)}}drawLine(e,t,r,i,o){let a=this.ctx,n={top:10,right:10,bottom:30,left:45},s=r-n.left-n.right,l=i-n.top-n.bottom,m=t.flatMap((d)=>d.data),u=Math.max(...m,0)*1.1||1;a.strokeStyle="rgba(63, 63, 70, 0.3)",a.lineWidth=0.5;for(let d=0;d<=4;d++){let c=n.top+l-d/4*l;a.beginPath(),a.moveTo(n.left,c),a.lineTo(r-n.right,c),a.stroke(),a.fillStyle="#71717a",a.font="10px -apple-system, sans-serif",a.textAlign="right",a.fillText(this.formatNum(u/4*d),n.left-5,c+3)}for(let d=0;d<e.length;d++){let c=n.left+d/(e.length-1||1)*s;a.fillStyle="#71717a",a.font="10px -apple-system, sans-serif",a.textAlign="center",a.fillText(this.truncLabel(e[d],10),c,i-8)}for(let d of t){a.strokeStyle=d.color,a.lineWidth=2,a.lineJoin="round",a.beginPath();let c=[];for(let f=0;f<d.data.length;f++){let g=n.left+f/(d.data.length-1||1)*s,v=n.top+l-d.data[f]/u*l;if(c.push([g,v]),f===0)a.moveTo(g,v);else a.lineTo(g,v)}if(a.stroke(),o&&c.length>0){a.globalAlpha=0.1,a.fillStyle=d.color,a.beginPath(),a.moveTo(c[0][0],n.top+l);for(let[f,g]of c)a.lineTo(f,g);a.lineTo(c[c.length-1][0],n.top+l),a.closePath(),a.fill(),a.globalAlpha=1}for(let[f,g]of c)a.fillStyle=d.color,a.beginPath(),a.arc(f,g,3,0,Math.PI*2),a.fill()}}drawPie(e,t,r){let i=this.ctx,o=e[0]?.data||[],a=o.reduce((u,d)=>u+d,0)||1,n=t/2,s=r/2,l=Math.min(n,s)-20,m=-Math.PI/2;for(let u=0;u<o.length;u++){let d=o[u]/a*Math.PI*2;i.fillStyle=P[u%P.length],i.beginPath(),i.moveTo(n,s),i.arc(n,s,l,m,m+d),i.closePath(),i.fill(),i.strokeStyle="var(--haira-bg-card, #0f0f12)",i.lineWidth=2,i.stroke(),m+=d}}drawScatter(e,t,r,i){this.drawLine(e,t,r,i,!1)}formatNum(e){if(e>=1e6)return`${(e/1e6).toFixed(1)}M`;if(e>=1000)return`${(e/1000).toFixed(1)}K`;return e%1===0?String(e):e.toFixed(1)}truncLabel(e,t){return e.length>t?e.slice(0,t-1)+"…":e}}var We={table:"haira-ui-table","status-card":"haira-ui-status-card","code-block":"haira-ui-code-block",diff:"haira-ui-diff","key-value":"haira-ui-key-value",progress:"haira-ui-progress-view",form:"haira-ui-form-view",confirm:"haira-ui-confirm",choices:"haira-ui-choices",chart:"haira-ui-chart"},Ye=3;class Ee extends HTMLElement{pendingEvent=null;connectedCallback(){if(this.ensureShadow(),this.pendingEvent)this.doRender(this.pendingEvent),this.pendingEvent=null}ensureShadow(){if(this.shadowRoot)return;this.attachShadow({mode:"open"}),this.shadowRoot.innerHTML=`
      <style>
        ${$}
        :host {
          display: block;
          margin-left: 2.25rem;
          max-width: 560px;
        }
        @media (max-width: 640px) {
          :host { margin-left: 0; max-width: 100%; }
        }
        .group { display: flex; flex-direction: column; gap: 0.5rem; }
        .fallback {
          background: var(--haira-bg-card); border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius); padding: 0.75rem 1rem;
          font-family: var(--haira-mono); font-size: 0.75rem;
          color: var(--haira-text-dim); white-space: pre-wrap; overflow-x: auto;
        }
      </style>
      <div id="container"></div>
    `}render(e){if(this.ensureShadow(),!this.isConnected){this.pendingEvent=e;return}this.doRender(e)}doRender(e){let t=this.shadowRoot.getElementById("container");t.innerHTML="";try{let r=this.renderNode(e.component,e.props,0);if(r)t.appendChild(r)}catch{let r=document.createElement("div");r.className="fallback",r.textContent=JSON.stringify(e.props,null,2),t.appendChild(r)}}renderNode(e,t,r){if(r>Ye)return null;if(e==="group"){let s=document.createElement("div");s.className="group";let l=t.children||[];for(let m of l){let{component:u,props:d}=m;if(u&&d){let c=this.renderNode(u,d,r+1);if(c)s.appendChild(c)}}return s}let i=We[e];if(!i){let s=document.createElement("div");return s.className="fallback",s.textContent=JSON.stringify(t,null,2),s}let o=document.createElement(i),n=this.hasAttribute("data-restored")?{...t,_restored:!0}:t;return Promise.resolve().then(()=>{o.setProps(n)}),o}}customElements.define("haira-field",de);customElements.define("haira-result",ce);customElements.define("haira-step",pe);customElements.define("haira-pipeline",he);customElements.define("haira-message",me);customElements.define("haira-tool-card",ge);customElements.define("haira-ui-status-card",ve);customElements.define("haira-ui-table",ye);customElements.define("haira-ui-code-block",xe);customElements.define("haira-ui-diff",we);customElements.define("haira-ui-key-value",$e);customElements.define("haira-ui-progress-view",ke);customElements.define("haira-ui-form-view",Se);customElements.define("haira-ui-confirm",Te);customElements.define("haira-ui-choices",Ce);customElements.define("haira-ui-chart",Le);customElements.define("haira-ui-renderer",Ee);customElements.define("haira-form",ue);customElements.define("haira-index",fe);customElements.define("haira-chat",be);customElements.define("haira-app",le);
