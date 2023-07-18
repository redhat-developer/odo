import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { DevfileGet200Response, GeneralSuccess, TelemetryResponse } from '../api-gen';

@Injectable({
  providedIn: 'root'
})
export class OdoapiService {

  private base = "/api/v1";

  constructor(private http: HttpClient) { }

  getDevfile(): Observable<DevfileGet200Response> {
    return this.http.get<DevfileGet200Response>(this.base+"/devfile");
  }

  saveDevfile(content: string): Observable<GeneralSuccess> {
    return this.http.put<GeneralSuccess>(this.base+"/devfile", {
      content: content
    });
  }

  telemetry(): Observable<TelemetryResponse> {
    return this.http.get<TelemetryResponse>(this.base+"/telemetry");
  }
}
