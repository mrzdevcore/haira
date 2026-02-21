// Backward compatibility — re-exports from core.
// Existing components import from here; new code should import from "core".

import { icons } from "./core/icons";

export { baseCSS as baseStyles, scrollbarCSS as scrollbarStyles, keyframes as sharedKeyframes, methodColor, uiTypeColor } from "./core/styles";
export { logoSvg } from "./core/icons";

export const iconPending = icons.pending;
export const iconSpinner = icons.spinner;
export const iconCheck = icons.check;
export const iconX = icons.x;
export const iconRetry = icons.retry;
export const iconCopy = icons.copy;
export const iconCopyDone = icons.copyDone;
export const iconChevron = icons.chevron;
