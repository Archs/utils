set nocompatible              " be iMproved, required
filetype off                  " required

" set the runtime path to include Vundle and initialize
set rtp+=~/.vim/bundle/vundle/
call vundle#rc()
" " alternatively, pass a path where Vundle should install bundles
" "let path = '~/some/path/here'
 "call vundle#rc(path)
"
" " let Vundle manage Vundle, required
Bundle 'gmarik/vundle'
"Bundle 'cespare/vim-golang'
Bundle 'desert256.vim'
Bundle 'ZenCoding.vim'
Bundle 'Blackrush/vim-gocode'
Bundle 'scrooloose/nerdtree'
Bundle 'majutsushi/tagbar'
Bundle 'ctrlp.vim'
Bundle 'Lokaltog/vim-powerline'
"Bundle 'minibufexpl.vim'
Bundle 'bufexplorer.zip'
Bundle 'SuperTab'
Bundle 'jiangmiao/auto-pairs'

" tag bar
"TagbarToggle
map <F9> :TagbarToggle <CR>
let g:tagbar_width = 25
let g:tagbar_type_go = {
    \ 'ctagstype' : 'go',
    \ 'kinds'     : [
        \ 'p:package',
        \ 'i:imports:1',
        \ 'c:constants',
        \ 'v:variables',
        \ 't:types',
        \ 'n:interfaces',
        \ 'w:fields',
        \ 'e:embedded',
        \ 'm:methods',
        \ 'r:constructor',
        \ 'f:functions'
    \ ],
    \ 'sro' : '.',
    \ 'kind2scope' : {
        \ 't' : 'ctype',
        \ 'n' : 'ntype'
    \ },
    \ 'scope2kind' : {
        \ 'ctype' : 't',
        \ 'ntype' : 'n'
    \ },
    \ 'ctagsbin'  : 'gotags',
    \ 'ctagsargs' : '-sort -silent'
    \ }

let g:gofmt_command = 'goimports'

""""""""setting of nerdtree
let NERDTreeWinPos='left'
let NERDTreeWinSize=25
let NERDTreeChDirMode=1
map <F8> :NERDTreeToggle <CR>
let NERDTreeDirArrows = 0

" other config
autocmd BufWritePre *.go :Fmt
" filetype plugin indent on
" syntax on

filetype plugin on

syntax on

autocmd BufReadPost *.cpp,*.c,*.h,*.hpp,*.cc,*.cxx,*.go,*.py call tagbar#autoopen()

" multi-encoding setting
if has("multi_byte")
"set bomb 
set fileencodings=ucs-bom,utf-8,cp936,gb18030,big5,euc-jp,sjis,euc-kr,ucs-2le,latin1 
" CJK environment detection and corresponding setting 
if v:lang =~ "^zh_CN" 
" Use cp936 to support GBK, euc-cn == gb2312 
set encoding=cp936 
set termencoding=cp936 
set fileencoding=cp936 
endif 
" Detect UTF-8 locale, and replace CJK setting if needed 
if v:lang =~ "utf8$" || v:lang =~ "UTF-8$" 
set encoding=utf-8 
set termencoding=utf-8 
set fileencoding=utf-8 
endif 
else 
echoerr "Sorry, this version of (g)vim was not compiled with multi_byte" 
endif 

"minibufexp
let g:miniBufExplMapWindowNavVim = 1

"关掉备份文件
set nobackup

"暂时没用，留着吧
hi CursorLine guibg=LightBlue

"高亮当前行
set cursorline

"显示行号
set number

"auto complete
let g:SuperTabRetainCompletionType = 2
let g:SuperTabDefaultCompletionType = "<C-X><C-O>" 

set autochdir
set clipboard=unnamed
set smarttab
set smartindent
colorscheme desert
set smartindent  
set smarttab
set expandtab  
set tabstop=4  
set softtabstop=4  
set shiftwidth=4  
set backspace=2
"set textwidth=79
set autoindent
set autochdir
set clipboard=unnamed
