<?php

namespace App\Http\Proxy;

class AdminProxy
{
    protected $http;


    /*
     *
     * 
     * 
     */
    public function __construct(\GuzzleHttp\Client $m_http)
    {
        $this->http = $m_http;
    }

    public function proxy($grantType, array $data = [])
    {
        $data = array_merge($data,[
            'grant_type' => $grantType,
            'client_id' => env('PASSPORT_Client_ID'),
            'client_secret' => env('PASSPORT_Client_SECRET'),
            'scope' => '', // env('OAUTH_SCOPE_PASSWORD'),
        ]);
        try {
            $response = $this->http->post(url('/oauth/token'), [
                'form_params' => $data,
            ]);

            
            return response(json_decode((string)$response->getBody()->getContents(), true))->cookie(
               'refreshToken', env('TOKEN_EXPIRE_TIME'), null, null, false,
               true // last , http only for defence xss attack 
            );
        } catch (\Exception $e) {
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
