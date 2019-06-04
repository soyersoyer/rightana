import { Validators } from '@angular/forms';

export class RValidators {
    static password = Validators.compose([Validators.required, Validators.minLength(8)]);
    static collectionName = Validators.compose([Validators.required, Validators.pattern("^[a-z0-9.]+$")]);
    static userName = Validators.compose([Validators.required, Validators.pattern("^[a-z0-9.]+$")]);
    static email = Validators.compose([Validators.required, Validators.email]);
}
