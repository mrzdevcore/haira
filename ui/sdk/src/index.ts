// Haira Console — entry point
// All components self-register via @customElement decorators.
// Import order matters: leaf components first, then containers.

// Generative UI components (leaf)
import "./components/haira-message";
import "./components/haira-tool-card";
import "./components/haira-field";
import "./components/haira-result";
import "./components/haira-step";
import "./components/haira-pipeline";
import "./components/haira-status-card";
import "./components/haira-table";
import "./components/haira-code-block";
import "./components/haira-diff";
import "./components/haira-key-value";
import "./components/haira-progress-view";
import "./components/haira-form-view";
import "./components/haira-confirm";
import "./components/haira-choices";
import "./components/haira-chart";
import "./components/haira-ui-renderer";

// Workbench components
import "./components/haira-chat";
import "./components/haira-form";

// Pages
import "./pages/home";
import "./pages/workbench";
import "./pages/workflows";
import "./pages/agents";
import "./pages/observe";
import "./pages/settings";

// App shell (must be last — reads metadata and creates the above)
import "./components/haira-app";
