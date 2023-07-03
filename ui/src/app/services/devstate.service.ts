import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';

type ChartResult = {
  chart: string;
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

@Injectable({
  providedIn: 'root'
})
export class DevstateService {

  private base = "/api/v1/devstate";

  constructor(private http: HttpClient) { }

  addContainer(container: Container): Observable<ResultValue> {
    return this.http.post<ResultValue>(this.base+"/container", {
      name: container.name,
      image: container.image,
      command: container.command,
      args: container.args,
      memReq: container.memoryRequest,
      memLimit: container.memoryLimit,
      cpuReq: container.cpuRequest,
      cpuLimit: container.cpuLimit,
    });
  }

  addImage(image: Image): Observable<ResultValue> {
    return this.http.post<ResultValue>(this.base+"/image", {
      name: image.name,
      imageName: image.imageName,
      args: image.args,
      buildContext: image.buildContext,
      rootRequired: image.rootRequired,
      uri: image.uri
    });
  }

  addResource(resource: ClusterResource): Observable<ResultValue> {
    return this.http.post<ResultValue>(this.base+"/resource", {
      name: resource.name,
      inlined: resource.inlined,
      uri: resource.uri,
    });
  }

  addExecCommand(name: string, cmd: ExecCommand): Observable<ResultValue> {
    return this.http.post<ResultValue>(this.base+"/execCommand", {
      name: name,
      component: cmd.component,
      commandLine: cmd.commandLine,
      workingDir: cmd.workingDir,
      hotReloadCapable: cmd.hotReloadCapable,
    });
  }

  addApplyCommand(name: string, cmd: ApplyCommand): Observable<ResultValue> {
    return this.http.post<ResultValue>(this.base+"/applyCommand", {
      name: name,
      component: cmd.component,
    });
  }

  addCompositeCommand(name: string, cmd: CompositeCommand): Observable<ResultValue> {
    return this.http.post<ResultValue>(this.base+"/compositeCommand", {
      name: name,
      parallel: cmd.parallel,
      commands: cmd.commands,
    });
  }

  // getFlowChart calls the wasm module to get the lifecycle of the Devfile in mermaid chart format
  getFlowChart(): Observable<ChartResult> {
    return this.http.get<ChartResult>(this.base+"/chart");
  }

  // setDevfileContent calls the wasm module to reset the content of the Devfile
  setDevfileContent(devfile: string): Observable<ResultValue> {
    return this.http.put<ResultValue>(this.base+"/devfile", {
      content: devfile
    });
  }

  // getDevfileContent gets the content of the Devfile
  getDevfileContent(): Observable<ResultValue> {
    return this.http.get<ResultValue>(this.base+"/devfile");
  }
  
  // clearDevfileContent clears the content of the Devfile
  clearDevfileContent(): Observable<ResultValue> {
    return this.http.delete<ResultValue>(this.base+"/devfile");
  }
  
  setMetadata(metadata: Metadata): Observable<ResultValue> {
    return this.http.put<ResultValue>(this.base+"/metadata", {
      name: metadata.name,
      version: metadata.version,
      displayName: metadata.displayName,
      description: metadata.description,
      tags: metadata.tags,
      architectures: metadata.architectures,
      icon: metadata.icon,
      globalMemoryLimit: metadata.globalMemoryLimit,
      projectType: metadata.projectType,
      language: metadata.language,
      website: metadata.website,
      provider: metadata.provider,
      supportUrl: metadata.supportUrl,
    });
  }

  moveCommand(previousKind: string, newKind: string, previousIndex: number, newIndex: number): Observable<ResultValue> {
    // TODO set correct command Name
    return this.http.post<ResultValue>(this.base+"/command/0/move", {
      fromGroup: previousKind,
      fromIndex: previousIndex,
      toGroup: newKind,
      toIndex: newIndex
    });
  }

  setDefaultCommand(command: string, group: string): Observable<ResultValue> {
    return this.http.post<ResultValue>(this.base+"/command/"+command+"/setDefault", {
      group: group
    });
  }

  unsetDefaultCommand(command: string): Observable<ResultValue> {
    return this.http.post<ResultValue>(this.base+"/command/"+command+"/unsetDefault", {});
  }

  deleteCommand(command: string): Observable<ResultValue>  {
    return this.http.delete<ResultValue>(this.base+"/command/"+command);
  }

  deleteContainer(container: string): Observable<ResultValue> {
    return this.http.delete<ResultValue>(this.base+"/container/"+container);
  }

  deleteImage(image: string): Observable<ResultValue> {
    return this.http.delete<ResultValue>(this.base+"/image/"+image);
  }

  deleteResource(resource: string): Observable<ResultValue> {
    return this.http.delete<ResultValue>(this.base+"/resource/"+resource);
  }

  updateEvents(event: "preStart"|"postStart"|"preStop"|"postStop", commands: string[]): Observable<ResultValue> {
    return this.http.put<ResultValue>(this.base+"/events", {
      eventName: event,
      commands: commands
    });
  }

  isQuantityValid(quantity: string): Observable<{}> {
    return this.http.post<{}>(this.base+"/quantityValid", {
      quantity: quantity
    });
  }
}
