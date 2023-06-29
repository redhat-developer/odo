import { Injectable } from '@angular/core';

type ChartResult = {
  err: string;
  value: any;
};

type Result = {
  err: string;
  value: ResultValue;
};

export type ResultValue = {
  content: string;
  metadata: Metadata;
  commands: Command[];
  events: Events;
  containers: Container[];
  images: Image[];
  resources: ClusterResource[];
};

export type Metadata = {
  name: string | null;
  version: string | null;
  displayName: string | null;
  description: string | null;
  tags: string | null;
  architectures: string | null;
  icon: string | null;
  globalMemoryLimit: string | null;
  projectType: string | null;
  language: string | null;
  website: string | null;
  provider: string | null;
  supportUrl: string | null;
};

export type Command = {
  name: string;
  group: string;
  default: boolean;
  type: "exec" | "apply" | "image" | "composite";
  exec: ExecCommand | undefined;
  apply: ApplyCommand | undefined;
  image: ImageCommand | undefined;
  composite: CompositeCommand | undefined;
};

export type Events = {
  preStart: string[];
  postStart: string[];
  preStop: string[];
  postStop: string[];
};

export type ExecCommand = {
  component: string;
  commandLine: string;
  workingDir: string;
  hotReloadCapable: boolean;
};

export type ApplyCommand = {
  component: string;
};

export type ImageCommand = {
  component: string;
};

export type CompositeCommand = {
  commands: string[];
  parallel: boolean;
};

export type Container = {
  name: string;
  image: string;
  command: string[];
  args: string[];
  memoryRequest: string;
  memoryLimit: string;
  cpuRequest: string;
  cpuLimit: string;
};

export type Image = {
  name: string;
  imageName: string;
  args: string[];
  buildContext: string;
  rootRequired: boolean;
  uri: string;
};

export type ClusterResource = {
  name: string;
  inlined: string;
  uri: string;
};

declare const addContainer: (name: string, image: string, command: string[], args: string[], memReq: string, memLimit: string, cpuReq: string, cpuLimit: string) => Result;
declare const addImage: (name: string, imageName: string, args: string[], buildContext: string, rootRequired: boolean, uri: string) => Result;
declare const addResource: (name: string, inlined: string, uri: string) => Result;
declare const addExecCommand: (name: string, component: string, commmandLine: string, workingDir: string, hotReloadCapable: boolean) => Result;
declare const addApplyCommand: (name: string, component: string) => Result;
declare const addCompositeCommand: (name: string, parallel: boolean, commands: string[]) => Result;
declare const getFlowChart: () => ChartResult;
declare const setDevfileContent: (devfile: string) => Result;
declare const setMetadata: (metadata: Metadata) => Result;
declare const moveCommand: (previousKind: string, newKind: string, previousIndex: number, newIndex: number) => Result;
declare const setDefaultCommand: (command: string, group: string) => Result;
declare const unsetDefaultCommand: (command: string) => Result;
declare const deleteCommand: (command: string) => Result;
declare const deleteContainer: (container: string) => Result;
declare const deleteImage: (image: string) => Result;
declare const deleteResource: (resource: string) => Result;
declare const updateEvents: (event: string, commands: string[]) => Result;
declare const isQuantityValid: (quantity: string) => Boolean;

@Injectable({
  providedIn: 'root'
})
// WasmGoService uses the wasm module. 
// The module manages a single instance of a Devfile
export class WasmGoService {

  addContainer(container: Container): Result {
    return addContainer(
      container.name,
      container.image,
      container.command,
      container.args,
      container.memoryRequest,
      container.memoryLimit,
      container.cpuRequest,
      container.cpuLimit,
    );
  }

  addImage(image: Image): Result {
    return addImage(
      image.name,
      image.imageName,
      image.args,
      image.buildContext,
      image.rootRequired,
      image.uri,
    );
  }

  addResource(resource: ClusterResource): Result {
    return addResource(
      resource.name,
      resource.inlined,
      resource.uri,
    );
  }

  addExecCommand(name: string, cmd: ExecCommand): Result {
    return addExecCommand(
      name,
      cmd.component,
      cmd.commandLine,
      cmd.workingDir,
      cmd.hotReloadCapable,
    );
  }

  addApplyCommand(name: string, cmd: ApplyCommand): Result {
    return addApplyCommand(
      name,
      cmd.component,      
    );
  }

  addCompositeCommand(name: string, cmd: CompositeCommand): Result {
    return addCompositeCommand(
      name,
      cmd.parallel,
      cmd.commands,      
    );
  }

  // getFlowChart calls the wasm module to get the lifecycle of the Devfile in mermaid chart format
  getFlowChart(): string {
    const result = getFlowChart();
    return result.value;
  }

  // setDevfileContent calls the wasm module to reset the content of the Devfile
  setDevfileContent(devfile: string): Result {
    const result = setDevfileContent(devfile);
    return result;  
  }

  setMetadata(metadata: Metadata): Result {
    return setMetadata(metadata);
  }

  moveCommand(previousKind: string, newKind: string, previousIndex: number, newIndex: number): Result {
    return moveCommand(previousKind, newKind, previousIndex, newIndex);
  }

  setDefaultCommand(command: string, group: string): Result {
    return setDefaultCommand(command, group);
  }

  unsetDefaultCommand(command: string): Result {
    return unsetDefaultCommand(command);
  }

  deleteCommand(command: string): Result {
    const result = deleteCommand(command);
    return result;
  }

  deleteContainer(container: string): Result {
    const result = deleteContainer(container);
    return result;
  }

  deleteImage(image: string): Result {
    const result = deleteImage(image);
    return result;
  }

  deleteResource(resource: string): Result {
    const result = deleteResource(resource);
    return result;
  }

  updateEvents(event: "preStart"|"postStart"|"preStop"|"postStop", commands: string[]): Result {
    return updateEvents(event, commands);
  }

  isQuantityValid(quantity: string): Boolean {
    return isQuantityValid(quantity);
  }
}
