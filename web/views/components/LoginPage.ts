
// components/LoginPage.ts

import { ThemeManager } from './ThemeManager';
import { ToastManager } from './ToastManager';

class LoginPage {
    private form: HTMLFormElement | null;
    private titleElement: HTMLElement | null;
    private submitButton: HTMLButtonElement | null;
    private toggleLink: HTMLElement | null;
    private confirmPasswordGroup: HTMLElement | null;
    private confirmPasswordInput: HTMLInputElement | null;
    private emailInput: HTMLInputElement | null;
    private callbackURL: HTMLInputElement;

    private isSignUpMode: boolean = false;

    constructor() {
        // Initialize global components
        ThemeManager.init();
        ToastManager.init(); // Initialize toast for potential messages

        // Find form elements
        this.callbackURL = (document.getElementById("callbackURL") as HTMLInputElement)
        this.form = document.getElementById('auth-form') as HTMLFormElement;
        this.titleElement = document.getElementById('auth-title');
        this.submitButton = document.getElementById('auth-submit-button') as HTMLButtonElement;
        this.toggleLink = document.getElementById('auth-toggle-link');
        this.confirmPasswordGroup = document.getElementById('confirm-password-group');
        this.confirmPasswordInput = document.getElementById('confirm-password') as HTMLInputElement;
        this.emailInput = document.getElementById('email-address') as HTMLInputElement;


        if (!this.form || !this.titleElement || !this.submitButton || !this.toggleLink || !this.confirmPasswordGroup || !this.confirmPasswordInput || !this.emailInput) {
            console.error("LoginPage: Could not find all required authentication form elements.");
            return;
        }

        this.bindEvents();
        this.updateUI(); // Set initial UI state
    }

    private bindEvents(): void {
        this.toggleLink?.addEventListener('click', (e) => {
            e.preventDefault();
            this.isSignUpMode = !this.isSignUpMode;
            this.updateUI();
        });

        // Optional: Clear errors when changing modes
        // this.toggleLink?.addEventListener('click', () => {
        //    const errorElement = document.getElementById('auth-error-message');
        //    if (errorElement) errorElement.textContent = '';
        // });
    }

    private updateUI(): void {
        if (!this.form || !this.titleElement || !this.submitButton || !this.toggleLink || !this.confirmPasswordGroup || !this.confirmPasswordInput || !this.emailInput) {
            return;
        }

        if (this.isSignUpMode) {
            this.titleElement.textContent = 'Create your account';
            this.submitButton.textContent = 'Sign up';
            this.toggleLink.textContent = 'Already have an account? Sign In';
            this.confirmPasswordGroup.classList.remove('hidden');
            this.confirmPasswordInput.required = true;
            this.form.action = '/auth/signup?callbackURL=' + this.callbackURL.value; // Point form action to signup endpoint
            this.emailInput.autocomplete = 'email';
            (this.form.querySelector('#password') as HTMLInputElement).autocomplete = 'new-password';
        } else {
            this.titleElement.textContent = 'Sign in to your account';
            this.submitButton.textContent = 'Sign in';
            this.toggleLink.textContent = 'Need an account? Sign Up';
            this.confirmPasswordGroup.classList.add('hidden');
            this.confirmPasswordInput.required = false;
            this.confirmPasswordInput.value = ''; // Clear confirm password field
            this.form.action = '/auth/login?callbackURL=' + this.callbackURL.value; // Point form action to signup endpoint
            this.emailInput.autocomplete = 'email';
             (this.form.querySelector('#password') as HTMLInputElement).autocomplete = 'current-password';
        }
    }

    public static init(): LoginPage {
       return new LoginPage();
    }
}

// Initialize the component when the DOM is fully loaded
document.addEventListener('DOMContentLoaded', () => {
    LoginPage.init();
});
