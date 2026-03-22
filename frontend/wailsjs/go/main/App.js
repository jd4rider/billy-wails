// @ts-check
// Wails v3 bindings — uses @wailsio/runtime Call API
import { Call } from '@wailsio/runtime';

export function AddMemory(arg1) {
  return Call.ByName("main.App.AddMemory", arg1);
}

export function DeleteMemory(arg1) {
  return Call.ByName("main.App.DeleteMemory", arg1);
}

export function GetConversations() {
  return Call.ByName("main.App.GetConversations");
}

export function GetMemories() {
  return Call.ByName("main.App.GetMemories");
}

export function GetMessages(arg1) {
  return Call.ByName("main.App.GetMessages", arg1);
}

export function GetPlatform() {
  return Call.ByName("main.App.GetPlatform");
}

export function GetStatus() {
  return Call.ByName("main.App.GetStatus");
}

export function ListModels() {
  return Call.ByName("main.App.ListModels");
}

export function NewConversation() {
  return Call.ByName("main.App.NewConversation");
}

export function OpenInstallPage() {
  return Call.ByName("main.App.OpenInstallPage");
}

export function PopOut() {
  return Call.ByName("main.App.PopOut");
}

export function SendMessage(arg1) {
  return Call.ByName("main.App.SendMessage", arg1);
}

export function SetActiveConversation(arg1) {
  return Call.ByName("main.App.SetActiveConversation", arg1);
}
