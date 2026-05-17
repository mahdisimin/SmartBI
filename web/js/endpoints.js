export const endpoints = [
  {
    id: 'healthcheck',
    name: 'Healthcheck',
    description: 'Verifies that the backend server is running.',
    method: 'GET',
    path: '/healthcheck',
    fields: [],
  },
  {
    id: 'register',
    name: 'Register User',
    description: 'Creates a user account with a name, phone number, and password.',
    method: 'POST',
    path: '/user/register',
    fields: [
      { name: 'user_name', label: 'User name', type: 'text', placeholder: 'Test User', required: true },
      { name: 'phone_number', label: 'Phone number', type: 'tel', placeholder: '09201008700', required: true },
      { name: 'password', label: 'Password', type: 'password', placeholder: '123456', required: true },
    ],
  },
  {
    id: 'login',
    name: 'Login User',
    description: 'Authenticates a user by phone number and password.',
    method: 'POST',
    path: '/user/login',
    fields: [
      { name: 'phone_number', label: 'Phone number', type: 'tel', placeholder: '09201008700', required: true },
      { name: 'password', label: 'Password', type: 'password', placeholder: '123456', required: true },
    ],
  },
];
