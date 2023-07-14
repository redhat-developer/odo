import { Injectable, NgZone } from '@angular/core';
import {Observable} from "rxjs";

@Injectable({
    providedIn: 'root'
})
export class SseService {
    private base = "/api/v1";
    private evtSource: EventSource

    constructor(private _zone: NgZone) {
        this.evtSource = new EventSource(this.base + "/notifications");
    }

    subscribeTo(eventTypes: string[]): Observable<any> {
        return new Observable( (subscriber) => {
            eventTypes.forEach(eventType => {
                this.evtSource.addEventListener(eventType,  (event) => {
                    this._zone.run(() => {
                        subscriber.next(event);
                    });
                });
            })
            this.evtSource.onerror = (error) => {
                this._zone.run(() => {
                    subscriber.error(error);
                });
            };
        });
    }
}
