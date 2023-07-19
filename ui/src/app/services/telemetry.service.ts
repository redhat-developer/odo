import { Injectable } from '@angular/core';
import { SegmentService } from 'ngx-segment-analytics';

@Injectable({
  providedIn: 'root'
})
export class TelemetryService {

  private options = {
    context: {
      ip: "0.0.0.0"
    }
  };

  constructor(
    private segment: SegmentService
  ) { }

  init(apikey: string, userid: string) {
    this.segment.identify(userid, {}, this.options);
    this.segment.load(apikey);
    this.segment.setAnonymousId(userid);
  }

  track(event: string) {
    this.segment.track(event, {}, this.options);
  }
}
