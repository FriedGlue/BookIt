import { token } from '$lib/stores/authStore';
import { PUBLIC_USER_POOL_CLIENT_ID, PUBLIC_USER_POOL_ID } from '$env/static/public';
import { CognitoUser, AuthenticationDetails, CognitoUserPool } from 'amazon-cognito-identity-js';

export class AuthService {
    private readonly userPool: CognitoUserPool;

    constructor() {
        this.userPool = new CognitoUserPool({
            UserPoolId: PUBLIC_USER_POOL_ID,
            ClientId: PUBLIC_USER_POOL_CLIENT_ID
        });
    }

    async login(username: string, password: string): Promise<void> {
        return new Promise((resolve, reject) => {
            const user = new CognitoUser({
                Username: username,
                Pool: this.userPool
            });

            const authDetails = new AuthenticationDetails({
                Username: username,
                Password: password
            });

            user.authenticateUser(authDetails, {
                onSuccess: (result) => {
                    const jwtToken = result.getIdToken().getJwtToken();
                    token.set(jwtToken);
                    window.location.href = '/';
                    resolve();
                },
                onFailure: (err) => {
                    reject(err);
                }
            });
        });
    }

    logout() {
        const currentUser = this.userPool.getCurrentUser();
        if (currentUser) {
            currentUser.signOut();
            token.set('');
            window.location.reload();
        }
    }
}