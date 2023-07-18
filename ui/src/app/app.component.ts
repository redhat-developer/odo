import { Component, OnInit } from '@angular/core';
import { DevstateService } from './services/devstate.service';
import { DomSanitizer } from '@angular/platform-browser';
import { MermaidService } from './services/mermaid.service';
import { StateService } from './services/state.service';
import { MatIconRegistry } from "@angular/material/icon";
import { OdoapiService } from './services/odoapi.service';
import { SseService } from './services/sse.service';
import {DevfileContent} from "./api-gen";

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnInit {

  protected mermaidContent: string = "";
  protected devfileYaml: string = "";
  protected errorMessage: string  = "";

  constructor(
    protected sanitizer: DomSanitizer,
    private matIconRegistry: MatIconRegistry,
    private wasmGo: DevstateService,
    private odoApi: OdoapiService,
    private mermaid: MermaidService,
    private state: StateService,
    private sse: SseService,
  ) {
    this.matIconRegistry.addSvgIcon(
      `github`,
      this.sanitizer.bypassSecurityTrustResourceUrl(`../assets/github-24.svg`)
    );
  }

  ngOnInit() {
    const loading = document.getElementById("loading");
    if (loading != null) {
      loading.style.visibility = "hidden";
    }

    const devfile = this.odoApi.getDevfile();
    devfile.subscribe({
      next: (devfile) => {
        if (devfile.content != undefined) {
          this.onButtonClick(devfile.content, false);
        }
      }
    });

    this.state.state.subscribe(async newContent => {
      if (newContent == null) {
        return;
      }

      this.devfileYaml = newContent.content;

      const result = this.wasmGo.getFlowChart();
      result.subscribe({
        next: async (res) => {
          const svg = await this.mermaid.getMermaidAsSVG(res.chart);
          this.mermaidContent = svg;      
        },
        error: (error) => {
          console.log(error);
        }
      });
    });

    this.sse.subscribeTo(['DevfileUpdated']).subscribe(event => {
      let newDevfile: DevfileContent = JSON.parse(event.data)
      if (newDevfile.content != undefined) {
        this.onButtonClick(newDevfile.content, false);
      }
    });
  }

  onButtonClick(content: string, save: boolean){
    const result = this.wasmGo.setDevfileContent(content);
    result.subscribe({
      next: (value) => {
        this.errorMessage = '';
        this.state.changeDevfileYaml(value);
        if (save) {
          this.odoApi.saveDevfile(value.content).subscribe({
            next: () => {},
            error: (error) => {
              this.errorMessage = error.error.message;
            }
          });
        }
      },
      error: (error) => {
        this.errorMessage = error.error.message;
      }
    });
  }

  clear() {
    if (confirm('You will delete the content of the Devfile. Continue?')) {
      this.wasmGo.clearDevfileContent().subscribe({
        next: (value) => {
          this.onButtonClick(value.content, false);
        }
      });
    }
  }
}
