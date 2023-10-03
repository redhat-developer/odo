import { Injectable } from '@angular/core';
import mermaid from 'mermaid';

@Injectable({
  providedIn: 'root'
})
export class MermaidService {

  constructor() { }

  async getMermaidAsSVG(definition: string): Promise<string> {
    const { svg } = await mermaid.render('rendered', definition);
    return svg;
  }
}
