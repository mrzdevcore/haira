import { HairaApp } from "./components/haira-app";
import { HairaField } from "./components/haira-field";
import { HairaResult } from "./components/haira-result";
import { HairaStep } from "./components/haira-step";
import { HairaPipeline } from "./components/haira-pipeline";
import { HairaMessage } from "./components/haira-message";
import { HairaForm } from "./components/haira-form";
import { HairaIndex } from "./components/haira-index";
import { HairaChat } from "./components/haira-chat";
import { HairaToolCard } from "./components/haira-tool-card";
import { HairaStatusCard } from "./components/haira-status-card";
import { HairaTable } from "./components/haira-table";
import { HairaCodeBlock } from "./components/haira-code-block";
import { HairaDiff } from "./components/haira-diff";
import { HairaKeyValue } from "./components/haira-key-value";
import { HairaProgressView } from "./components/haira-progress-view";
import { HairaFormView } from "./components/haira-form-view";
import { HairaConfirm } from "./components/haira-confirm";
import { HairaChoices } from "./components/haira-choices";
import { HairaUIRenderer } from "./components/haira-ui-renderer";

// Register leaf components first — container components may create children
// during connectedCallback, so children must already be defined.
customElements.define("haira-field", HairaField);
customElements.define("haira-result", HairaResult);
customElements.define("haira-step", HairaStep);
customElements.define("haira-pipeline", HairaPipeline);
customElements.define("haira-message", HairaMessage);
customElements.define("haira-tool-card", HairaToolCard);
customElements.define("haira-ui-status-card", HairaStatusCard);
customElements.define("haira-ui-table", HairaTable);
customElements.define("haira-ui-code-block", HairaCodeBlock);
customElements.define("haira-ui-diff", HairaDiff);
customElements.define("haira-ui-key-value", HairaKeyValue);
customElements.define("haira-ui-progress-view", HairaProgressView);
customElements.define("haira-ui-form-view", HairaFormView);
customElements.define("haira-ui-confirm", HairaConfirm);
customElements.define("haira-ui-choices", HairaChoices);
customElements.define("haira-ui-renderer", HairaUIRenderer);
customElements.define("haira-form", HairaForm);
customElements.define("haira-index", HairaIndex);
customElements.define("haira-chat", HairaChat);
// haira-app must be last — it reads metadata and creates the above components
customElements.define("haira-app", HairaApp);
