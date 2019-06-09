<?php

namespace App\Http\Controllers\back_api;

use Illuminate\Http\Request;
use App\Http\Controllers\Controller;
use Auth;

class AdminController extends Controller
{
    public function login(Request $request)
    {
        $http = new \GuzzleHttp\Client();
        try
        {
            $response = $http->post(url('/oauth/token'), [
                'form_params' => [
                    'grant_type' => env('OAUTH_GRANT_TYPE_PASSWORD'),
                    'client_id' => env('PASSPORT_Client_ID'),
                    'client_secret' => env('PASSPORT_Client_SECRET'),
                    'scope' => '', // env('OAUTH_SCOPE_PASSWORD'),
                    'username' => $request->input('username'),
                    'password' => $request->input('password'),
                ],
            ]);
            return json_decode((string)$response->getBody()->getContents(), true);
        } catch (RequestException $e) {
            if ($e->hasResponse()) {
                return $e->getResponse();
            }
            $code = $e->getCode();
            $message = 'ERROR:';
            $message .= $code;
    
            return response()->json([
                'code' => 200,  // or $code
                'message' => $message
            ], 200 /* or $code */);
        };
    }
}
