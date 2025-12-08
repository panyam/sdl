import { vi } from 'vitest';

export const editor = {
  create: vi.fn(() => ({
    dispose: vi.fn(),
    onDidChangeModelContent: vi.fn(),
    setValue: vi.fn(),
    getValue: vi.fn(() => ''),
    getModel: vi.fn(() => ({
      getValue: vi.fn(() => ''),
      getLineCount: vi.fn(() => 1),
      getLineMaxColumn: vi.fn(() => 1)
    })),
    focus: vi.fn(),
    layout: vi.fn(),
    deltaDecorations: vi.fn(() => []),
    revealLineInCenter: vi.fn()
  })),
  setTheme: vi.fn(),
  defineTheme: vi.fn(),
  setModelMarkers: vi.fn(),
  createModel: vi.fn(() => ({
    dispose: vi.fn(),
    getValue: vi.fn(() => ''),
    setValue: vi.fn()
  }))
};

export const languages = {
  register: vi.fn(),
  setMonarchTokensProvider: vi.fn(),
  setLanguageConfiguration: vi.fn()
};

export const MarkerSeverity = {
  Error: 8,
  Warning: 4,
  Info: 2,
  Hint: 1
};

export const KeyMod = {
  CtrlCmd: 2048,
  Shift: 1024,
  Alt: 512,
  WinCtrl: 256
};

export const KeyCode = {
  Enter: 3,
  Space: 10,
  Tab: 2,
  Escape: 9,
  Backspace: 1,
  Delete: 20,
  F1: 59,
  F2: 60,
  KeyS: 83
};

export default {
  editor,
  languages,
  MarkerSeverity,
  KeyMod,
  KeyCode
};