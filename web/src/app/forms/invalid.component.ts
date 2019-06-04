import { Component, OnInit } from '@angular/core';

@Component({
  selector: 'rana-invalid-username',
  template: `Usernames can contain letters (a-z), numbers (0-9), and dot (.).`,
})
export class InvalidUsernameComponent {}

@Component({
    selector: 'rana-invalid-password',
    template: `Your password must be at least 8 characters long.`,
})
export class InvalidPasswordComponent {}

@Component({
    selector: 'rana-invalid-email',
    template: `Please enter a valid email address!`,
})
export class InvalidEmailComponent {}

@Component({
    selector: 'rana-invalid-collection-name',
    template: `Collection names can contain letters (a-z), numbers (0-9), and dot (.).`,
  })
  export class InvalidCollectionNameComponent {}
  