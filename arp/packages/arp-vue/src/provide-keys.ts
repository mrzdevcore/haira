import type { InjectionKey } from "vue";
import type { ArpCapabilities } from "@haira/arp";

/** Injection key for ARP component registry. */
export const ArpComponentsKey: InjectionKey<Record<string, any>> = Symbol("arp-components");

/** Injection key for ARP connection state. */
export const ArpConnectedKey: InjectionKey<boolean> = Symbol("arp-connected");

/** Injection key for ARP capabilities. */
export const ArpCapabilitiesKey: InjectionKey<ArpCapabilities | null> = Symbol("arp-capabilities");
