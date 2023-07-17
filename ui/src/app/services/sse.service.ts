import { Injectable } from '@angular/core';
import {Observable} from "rxjs";

@Injectable({
    providedIn: 'root'
})
export class SseService {
    private base = "/api/v1";
    private evtSource: EventSource

    constructor() {
        this.evtSource = new EventSource(this.base + "/notifications");
    }

    subscribeTo(eventTypes: string[]): Observable<any> {
        return new Observable( (subscriber) => {
            eventTypes.forEach(eventType => {
                this.evtSource.addEventListener(eventType,  (event) => {
                    subscriber.next(event);
                });
            })
            this.evtSource.onerror = (error) => {
                subscriber.error(error);
            };
        });
    }
}
