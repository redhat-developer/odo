import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable, catchError, map, of } from 'rxjs';
import { ApplyCommand, CompositeCommand, Container, DevfileContent, DevstateChartGet200Response, ExecCommand, Image, Metadata, Resource, Volume } from '../api-gen';
import { AbstractControl, AsyncValidatorFn, ValidationErrors } from '@angular/forms';

@Injectable({
  providedIn: 'root'
})
export class DevstateService {

  private base = "/api/v1/devstate";

  constructor(private http: HttpClient) { }

  addContainer(container: Container): Observable<DevfileContent> {
    return this.http.post<DevfileContent>(this.base+"/container", {
      name: container.name,
      image: container.image,
      command: container.command,
      args: container.args,
      memReq: container.memoryRequest,
      memLimit: container.memoryLimit,
      cpuReq: container.cpuRequest,
      cpuLimit: container.cpuLimit,
      volumeMounts: container.volumeMounts,
      configureSources: container.configureSources,
      mountSources: container.mountSources,
      sourceMapping: container.sourceMapping,
    });
  }

  addImage(image: Image): Observable<DevfileContent> {
    return this.http.post<DevfileContent>(this.base+"/image", {
      name: image.name,
      imageName: image.imageName,
      args: image.args,
      buildContext: image.buildContext,
      rootRequired: image.rootRequired,
      uri: image.uri
    });
  }

  addResource(resource: Resource): Observable<DevfileContent> {
    return this.http.post<DevfileContent>(this.base+"/resource", {
      name: resource.name,
      inlined: resource.inlined,
      uri: resource.uri,
    });
  }

  addVolume(volume: Volume): Observable<DevfileContent> {
    return this.http.post<DevfileContent>(this.base+"/volume", {
      name: volume.name,
      ephemeral: volume.ephemeral,
      size: volume.size,
    });
  }

  addExecCommand(name: string, cmd: ExecCommand): Observable<DevfileContent> {
    return this.http.post<DevfileContent>(this.base+"/execCommand", {
      name: name,
      component: cmd.component,
      commandLine: cmd.commandLine,
      workingDir: cmd.workingDir,
      hotReloadCapable: cmd.hotReloadCapable,
    });
  }

  addApplyCommand(name: string, cmd: ApplyCommand): Observable<DevfileContent> {
    return this.http.post<DevfileContent>(this.base+"/applyCommand", {
      name: name,
      component: cmd.component,
    });
  }

  addCompositeCommand(name: string, cmd: CompositeCommand): Observable<DevfileContent> {
    return this.http.post<DevfileContent>(this.base+"/compositeCommand", {
      name: name,
      parallel: cmd.parallel,
      commands: cmd.commands,
    });
  }

  // getFlowChart calls the wasm module to get the lifecycle of the Devfile in mermaid chart format
  getFlowChart(): Observable<DevstateChartGet200Response> {
    return this.http.get<DevstateChartGet200Response>(this.base+"/chart");
  }

  // setDevfileContent calls the wasm module to reset the content of the Devfile
  setDevfileContent(devfile: string): Observable<DevfileContent> {
    return this.http.put<DevfileContent>(this.base+"/devfile", {
      content: devfile
    });
  }

  // getDevfileContent gets the content of the Devfile
  getDevfileContent(): Observable<DevfileContent> {
    return this.http.get<DevfileContent>(this.base+"/devfile");
  }
  
  // clearDevfileContent clears the content of the Devfile
  clearDevfileContent(): Observable<DevfileContent> {
    return this.http.delete<DevfileContent>(this.base+"/devfile");
  }
  
  setMetadata(metadata: Metadata): Observable<DevfileContent> {
    return this.http.put<DevfileContent>(this.base+"/metadata", {
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

  moveCommand(previousKind: string, newKind: string, previousIndex: number, newIndex: number): Observable<DevfileContent> {
    // TODO set correct command Name
    return this.http.post<DevfileContent>(this.base+"/command/0/move", {
      fromGroup: previousKind,
      fromIndex: previousIndex,
      toGroup: newKind,
      toIndex: newIndex
    });
  }

  setDefaultCommand(command: string, group: string): Observable<DevfileContent> {
    return this.http.post<DevfileContent>(this.base+"/command/"+command+"/setDefault", {
      group: group
    });
  }

  unsetDefaultCommand(command: string): Observable<DevfileContent> {
    return this.http.post<DevfileContent>(this.base+"/command/"+command+"/unsetDefault", {});
  }

  deleteCommand(command: string): Observable<DevfileContent>  {
    return this.http.delete<DevfileContent>(this.base+"/command/"+command);
  }

  deleteContainer(container: string): Observable<DevfileContent> {
    return this.http.delete<DevfileContent>(this.base+"/container/"+container);
  }

  deleteImage(image: string): Observable<DevfileContent> {
    return this.http.delete<DevfileContent>(this.base+"/image/"+image);
  }

  deleteResource(resource: string): Observable<DevfileContent> {
    return this.http.delete<DevfileContent>(this.base+"/resource/"+resource);
  }

  deleteVolume(volume: string): Observable<DevfileContent> {
    return this.http.delete<DevfileContent>(this.base+"/volume/"+volume);
  }

  updateEvents(event: "preStart"|"postStart"|"preStop"|"postStop", commands: string[]): Observable<DevfileContent> {
    return this.http.put<DevfileContent>(this.base+"/events", {
      eventName: event,
      commands: commands
    });
  }

  isQuantityValid(quantity: string): Observable<{}> {
    return this.http.post<{}>(this.base+"/quantityValid", {
      quantity: quantity
    });
  }

  isQuantity():  AsyncValidatorFn {
    return (control: AbstractControl): Observable<ValidationErrors | null> => {
      const val = control.value;
      if (val == '') {
        return of(null);
      }
      const valid = this.isQuantityValid(val);
      return valid.pipe(
        map(() => null),
        catchError(() => of({"isQuantity": false}))
      );
    };
  }  
}
