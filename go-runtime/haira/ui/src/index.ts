import { HairaApp } from "./components/haira-app";
import { HairaField } from "./components/haira-field";
import { HairaResult } from "./components/haira-result";
import { HairaStep } from "./components/haira-step";
import { HairaPipeline } from "./components/haira-pipeline";
import { HairaMessage } from "./components/haira-message";
import { HairaForm } from "./components/haira-form";
import { HairaIndex } from "./components/haira-index";
import { HairaChat } from "./components/haira-chat";

// Register leaf components first — container components may create children
// during connectedCallback, so children must already be defined.
customElements.define("haira-field", HairaField);
customElements.define("haira-result", HairaResult);
customElements.define("haira-step", HairaStep);
customElements.define("haira-pipeline", HairaPipeline);
customElements.define("haira-message", HairaMessage);
customElements.define("haira-form", HairaForm);
customElements.define("haira-index", HairaIndex);
customElements.define("haira-chat", HairaChat);
// haira-app must be last — it reads metadata and creates the above components
customElements.define("haira-app", HairaApp);
