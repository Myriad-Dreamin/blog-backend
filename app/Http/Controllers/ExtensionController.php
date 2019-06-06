<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
// use ___PHPSTORM_HELPERS\object;

class ExtensionController extends Controller
{
    public function secretToMe()
    {
        return view("extensions.secretlove");
    }

    public function chocolate()
    {
        $chocos = [
            (object) [
                'title' => 'Choco Giving',
                'category' => 'Any',
                'intro' => '要不要也给我推荐一点强大的工具QwQ',
                'ref' => 'http://myriaddreamin.com/secretlove'
            ],
            (object) [
                'title' => 'Git',
                'category' => 'Code',
                'intro' => 'Git包管理工具真的很好用呀QwQ',
                'ref' => 'https://git-scm.com/'
            ],
            (object) [
                'title' => 'Git Desktop',
                'category' => 'Code',
                'intro' => '用过Windows10版本的Git GUI吗',
                'ref' => 'https://desktop.github.com/'
            ],
            (object) [
                'title' => 'Visual Studio Code',
                'category' => 'Code',
                'intro' => '我身边的同学们都在用',
                'ref' => 'https://code.visualstudio.com/'
            ],
            (object) [
                'title' => 'php7',
                'category' => 'Language',
                'intro' => '世界上最好的语言',
                'ref' => 'https://www.php.net/'
            ],
            (object) [
                'title' => 'Python3',
                'category' => 'Language',
                'intro' => '世界上第二好的语言',
                'ref' => 'https://www.python.org/'
            ],
            (object) [
                'title' => 'Matplotlib',
                'category' => 'Pacakge',
                'intro' => '图形绘画工具',
                'ref' => 'https://matplotlib.org/'
            ]
        ];
        return view("extensions.chocolate", compact('chocos'));
    }

    public function musical()
    {
        // 0 "netease"
        $musics = [
            (object) [
                "recommand_type" => 0,
                "category" => "Traversal",
                "name" => "BLADE (satella Remix)",
                "artist" => "MAYA AKAI",
                "track" => "VIOLET",
                "ref" => "1365389564",
                "comment" => "某人推荐的qwq",
            ],
            (object) [
                "recommand_type" => 0,
                "category" => "Ejection",
                "name" => "Rainbow",
                "artist" => "DJ DiA/KO3",
                "track" => "Rainbow",
                "ref" => "1321385678",
                "comment" => "no response",
            ],
            (object) [
                "recommand_type" => 0,
                "category" => "Happystyle",
                "name" => "Trapped",
                "artist" => "Neilio/Miss Judged",
                "track" => "Trapped",
                "ref" => "32619803",
                "comment" => "no response",
            ],
            (object) [
                "recommand_type" => 0,
                "category" => "Tsuioku",
                "name" => "Traces of Pain - 伤痕",
                "artist" => "Eric Chiryoku",
                "track" => "The Beginning - 重新开始",
                "ref" => "1365275521",
                "comment" => "偶然遇见的，感觉还不错~|For HigHwind",
            ],
            (object) [
                "recommand_type" => 0,
                "category" => "Cluster",
                "name" => "The Void",
                "artist" => "月代彩",
                "track" => "fig.3: wave",
                "ref" => "477992821",
                "comment" => "no response|For Tan",
            ],
            (object) [
                "recommand_type" => 0,
                "category" => "Psychotonic",
                "name" => "Too Young to Die",
                "artist" => "polaritia",
                "track" => "SPRING 2019",
                "ref" => "1367532920",
                "comment" => "orz",
            ]
        ];

        foreach($musics as $mmusic) {
            switch ($mmusic->recommand_type) {
                case 0:
                    $mmusic->ref = "http://music.163.com/song/media/outer/url?id=". $mmusic->ref .".mp3";
                    break;
                default:
                dd("居然发现了错误信息..." . json_encode($mmusic));
            }
        }

        return view("extensions.musical", compact('musics'));
    }
}
