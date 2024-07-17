// Code generated - EDITING IS FUTILE. DO NOT EDIT.
//
// Generated by:
//     public/app/plugins/gen.go
// Using jennies:
//     TSTypesJenny
//     PluginTsTypesJenny
//
// Run 'make gen-cue' from repository root to regenerate.

import * as common from '@grafana/schema';

export const pluginVersion = "11.0.0";

export interface Options extends common.SingleStatBaseOptions {
  colorMode: common.BigValueColorMode;
  graphMode: common.BigValueGraphMode;
  justifyMode: common.BigValueJustifyMode;
  showPercentChange: boolean;
  textMode: common.BigValueTextMode;
  wideLayout: boolean;
}

export const defaultOptions: Partial<Options> = {
  colorMode: common.BigValueColorMode.Value,
  graphMode: common.BigValueGraphMode.Area,
  justifyMode: common.BigValueJustifyMode.Auto,
  showPercentChange: false,
  textMode: common.BigValueTextMode.Auto,
  wideLayout: true,
};
