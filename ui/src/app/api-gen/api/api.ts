export * from './default.service';
import { DefaultService } from './default.service';
export * from './devstate.service';
import { DevstateService } from './devstate.service';
export const APIS = [DefaultService, DevstateService];
